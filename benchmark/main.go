package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// 配置项
const (
	BaseURL    = "http://localhost:8080/api/v1"
	TotalUsers = 10000000 // 模拟用户数
	ProductID  = 101  // 商品ID
)

// 统计计数器
var (
	registerOk  int32
	loginOk     int32
	requestSent int32
	flashOk     int32 // 抢购成功 (202)
	soldOut     int32 // 售罄/重复 (400)
	fail        int32 // 系统错误
)

// 响应结构体用于解析 Token
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func main() {
	fmt.Printf("🚀 开始准备 %d 个用户环境...\n", TotalUsers)

	// 使用 waitGroup 等待所有用户准备就绪
	var wg sync.WaitGroup
	wg.Add(TotalUsers)

	// 用于控制“同时开抢”的闸门
	startGun := make(chan struct{})

	// 存放准备好 Token 的用户
	type Player struct {
		UserID int
		Token  string
	}
	players := make(chan Player, TotalUsers)

	// 1. 启动 10000000 个协程进行注册和登录 (蓄水阶段)
	for i := 1; i <= TotalUsers; i++ {
		go func(uid int) {
			defer wg.Done()
			username := fmt.Sprintf("user_%d_%d", time.Now().Unix(), uid)
			password := "123456"

			// --- Step A: 注册 ---
			if err := register(username, password); err != nil {
				fmt.Printf("用户 %d 注册失败: %v\n", uid, err)
				return
			}
			atomic.AddInt32(&registerOk, 1)

			// --- Step B: 登录 ---
			token, err := login(username, password)
			if err != nil {
				fmt.Printf("用户 %d 登录失败: %v\n", uid, err)
				return
			}
			atomic.AddInt32(&loginOk, 1)

			// 将拿到的 Token 放入通道，准备抢购
			players <- Player{UserID: uid, Token: token}
		}(i)
	}

	// 等待所有用户注册登录完成
	wg.Wait()
	close(players)
	fmt.Printf("✅ 环境准备完成！注册成功: %d, 登录成功: %d\n", registerOk, loginOk)
	fmt.Println("------------------------------------------------")
	fmt.Println("🔥 3秒后开始 10000000 人并发秒杀！Ready...")
	time.Sleep(1 * time.Second)
	fmt.Println("3...")
	time.Sleep(1 * time.Second)
	fmt.Println("2...")
	time.Sleep(1 * time.Second)
	fmt.Println("1... GO!!! 🔫")

	// 2. 开闸！并发抢购 (攻击阶段)
	var attackWg sync.WaitGroup
	// 重新读取 players 通道里的用户
	validPlayers := make([]Player, 0, TotalUsers)
	for p := range players {
		validPlayers = append(validPlayers, p)
	}

	attackWg.Add(len(validPlayers))

	startTime := time.Now()

	for _, p := range validPlayers {
		go func(player Player) {
			defer attackWg.Done()
			// 等待发令枪响 (实际上这里已经直接开始了，因为下面 close(startGun) 瞬间执行)
			<-startGun 
			
			// --- Step C: 抢购 ---
			code := flashBuy(player.Token, ProductID)
			atomic.AddInt32(&requestSent, 1)

			switch code {
			case 20200: // Accepted (抢购成功，排队中)
				atomic.AddInt32(&flashOk, 1)
			case 40010, 40001: // 售罄 或 重复
				atomic.AddInt32(&soldOut, 1)
			default:
				atomic.AddInt32(&fail, 1)
				fmt.Printf("用户 %d 异常: Code %d\n", player.UserID, code)
			}
		}(p)
	}

	// 瞬间开启闸门
	close(startGun)
	
	// 等待所有抢购请求结束
	attackWg.Wait()
	duration := time.Since(startTime)

	// 3. 输出统计报告
	fmt.Println("------------------------------------------------")
	fmt.Printf("🏁 压测结束！耗时: %v\n", duration)
	fmt.Printf("📊 总请求数: %d\n", requestSent)
	fmt.Printf("✅ 抢购受理 (HTTP 202): %d\n", flashOk)
	fmt.Printf("🚫 售罄/被拦截 (HTTP 400): %d\n", soldOut)
	fmt.Printf("❌ 异常失败: %d\n", fail)
	fmt.Printf("🚀 QPS: %.2f\n", float64(requestSent)/duration.Seconds())
	fmt.Println("------------------------------------------------")
	fmt.Println("请检查 Redis (flash:stock:101) 剩余库存是否正确 (预期0或剩余少量)")
	fmt.Println("请检查 MySQL (orders表) 订单数是否等于 50 (不超卖)")
}

// ---------------- 辅助函数 ----------------

func register(username, password string) error {
	body := map[string]string{"username": username, "password": password}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(BaseURL+"/users/register", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	return nil
}

func login(username, password string) (string, error) {
	body := map[string]string{"username": username, "password": password}
	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(BaseURL+"/users/login", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code %d", resp.StatusCode)
	}

	var res Response
	// 注意：根据你的代码 response.Success 返回的数据结构，Token 可能在 Data Map 里
	// 这里假设返回结构是 {"code": 20000, "data": {"access_token": "...", ...}}
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &res)

	dataMap, ok := res.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("response data format error")
	}
	return dataMap["access_token"].(string), nil
}

func flashBuy(token string, pid int) int {
	body := map[string]int{"product_id": pid}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", BaseURL+"/orders/flash-buy", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	// 解析返回的 Code (业务状态码)
	var res Response
	respBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBytes, &res)
	return res.Code
}