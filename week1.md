# Week 1
## 以太坊白皮书
---
1. 以太坊的优势
- 以太坊旨在提供一个内置图灵完备编程语言的区块链平台，允许用户创建任意规则的智能合约和去中心化应用（DApp），实现各类复杂的状态转换逻辑，覆盖金融、身份验证、文件存储等多种场景。
2. 以太坊跟别的货币的比较
- 比特币作为去中心化货币，其底层区块链技术有创新，但脚本语言存在局限：缺乏图灵完备性（无法实现循环等复杂逻辑）、价值盲（难以精细控制资金分配）、无状态（无法支持多阶段合约）、区块链盲（无法访问链上数据如时间戳）。
- 现有其他区块链应用（如域名币、彩色币、元币）存在开发成本高、安全性不足或扩展性差等问题。
3. 以太坊的核心技术
- 账户与状态：以太坊状态由“账户”组成，每个账户包含nonce（防止交易重放）、以太币余额、合约代码（若为合约账户）和存储。状态转换是账户间价值和信息的直接转移。
- 消息与交易：交易是外部账户发送的签名数据包，包含接收者、签名、转账金额、数据字段及燃料相关参数（ STARTGAS 和 GASPRICE ），用于限制计算资源消耗并支付费用；合约可发送“消息”，类似交易但由合约触发。
- 状态转换函数：定义了交易执行后状态的更新规则，包括验证交易有效性、扣除费用、执行合约代码（若接收者为合约）、处理燃料消耗及费用分配等。
- 代码执行：通过以太坊虚拟机（EVM）运行基于堆栈的字节码，支持访问堆栈、内存和长期存储，可处理复杂逻辑。
- 区块链与挖矿：区块包含交易列表、状态副本等信息，采用改进的GHOST协议处理区块分叉，激励矿工验证完整区块，同时通过燃料机制限制区块资源消耗。
4. 以太坊的应用场景
- 金融应用：如子货币、金融衍生品、对冲合约、储蓄钱包等。
- 半金融应用：如自动执行的计算难题赏金。
- 非金融应用：如在线投票、去中心化治理、身份与信誉系统、去中心化文件存储、去中心化自治组织（DAO）等。
5. 其他
- 燃料与费用：交易需支付燃料费用，防止恶意消耗资源，燃料成本与计算步骤、数据存储等相关。
- 以太币发行：通过货币销售初始发行，之后每年为矿工分配固定数量，长期供应增长率趋于零。
- 可扩展性与安全性：采用帕特里夏树优化状态存储，通过验证协议和质询-应答机制应对中心化风险，
---
## go语言学习
1. go语言介绍
- go语言是由谷歌在2009年开发的一款静态强类型、编译型编程语言。
2. go语言特点
- 简洁易学：语法简洁，摒弃了许多复杂特性（如继承、泛型早期不支持等），类似C语言风格，上手门槛低。
​
- 高效性能：编译速度快，执行效率接近C语言，适合开发高性能应用。
​
- 原生支持并发：通过goroutine（轻量级线程）和channel（通道）机制，轻松实现高并发编程，资源消耗低。
3. go语言基本知识点
``` go
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

// 交易结构
type Transaction struct {
	ID     string `json:"id"`
	Sender string `json:"sender"`   // 发送者地址
	Recipient string `json:"recipient"` // 接收者地址
	Amount  int    `json:"amount"`  // 金额
}

// 区块结构
type Block struct {
	Index     int           `json:"index"`
	Timestamp string        `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash  string        `json:"prev_hash"`
	Hash      string        `json:"hash"`
	Nonce     int           `json:"nonce"`     // 用于POW的随机数
	Difficulty int          `json:"difficulty"`// 挖矿难度
}

// 区块链结构
type Blockchain struct {
	blocks []*Block
}

// 生成交易ID
func generateTransactionID(tx Transaction) string {
	data, _ := json.Marshal(tx)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// 计算区块哈希
func calculateHash(block Block) string {
	blockData, _ := json.Marshal(block)
	hash := sha256.Sum256(blockData)
	return hex.EncodeToString(hash[:])
}

// 工作量证明：寻找符合难度的哈希
func proofOfWork(block *Block) {
	target := make([]byte, block.Difficulty)
	for i := range target {
		target[i] = '0'
	}
	targetStr := string(target)

	for {
		block.Hash = calculateHash(*block)
		if block.Hash[:block.Difficulty] == targetStr {
			break
		}
		block.Nonce++
	}
}

// 创建新交易
func (bc *Blockchain) createTransaction(sender, recipient string, amount int) Transaction {
	tx := Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
	}
	tx.ID = generateTransactionID(tx)
	return tx
}

// 创建新区块
func (bc *Blockchain) createBlock(transactions []Transaction, difficulty int) *Block {
	var prevHash string
	if len(bc.blocks) == 0 {
		prevHash = "" // 创世区块无前哈希
	} else {
		prevHash = bc.blocks[len(bc.blocks)-1].Hash
	}

	block := &Block{
		Index:       len(bc.blocks),
		Timestamp:   time.Now().String(),
		Transactions: transactions,
		PrevHash:    prevHash,
		Difficulty:  difficulty,
		Nonce:       0,
	}

	// 执行工作量证明
	proofOfWork(block)
	return block
}

// 添加区块到链
func (bc *Blockchain) addBlock(block *Block) {
	if len(bc.blocks) == 0 {
		// 验证创世区块
		bc.blocks = append(bc.blocks, block)
		return
	}

	// 验证新区块（哈希和前哈希匹配）
	lastBlock := bc.blocks[len(bc.blocks)-1]
	if block.PrevHash == lastBlock.Hash && block.Hash == calculateHash(*block) {
		bc.blocks = append(bc.blocks, block)
	} else {
		log.Println("Invalid block, cannot add to chain")
	}
}

// 打印区块链
func (bc *Blockchain) printChain() {
	for _, block := range bc.blocks {
		fmt.Printf("\nBlock #%d\n", block.Index)
		fmt.Printf("Timestamp: %s\n", block.Timestamp)
		fmt.Printf("Transactions: %+v\n", block.Transactions)
		fmt.Printf("Prev Hash: %s\n", block.PrevHash)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Printf("Difficulty: %d\n", block.Difficulty)
	}
}

// 初始化区块链（创建创世区块）
func initBlockchain() *Blockchain {
	bc := &Blockchain{}
	// 创建创世区块（包含一笔系统奖励交易）
	genesisTx := Transaction{
		ID:        "genesis",
		Sender:    "system",
		Recipient: "miner-initial",
		Amount:    100, // 初始奖励
	}
	genesisBlock := bc.createBlock([]Transaction{genesisTx}, 4) // 难度4
	bc.addBlock(genesisBlock)
	return bc
}

func main() {
	// 初始化区块链
	bc := initBlockchain()
	fmt.Println("=== Blockchain after genesis block ===")
	bc.printChain()

	// 创建一些交易
	tx1 := bc.createTransaction("alice", "bob", 5)
	tx2 := bc.createTransaction("bob", "charlie", 3)

	// 挖矿创建第二个区块（难度4）
	block2 := bc.createBlock([]Transaction{tx1, tx2}, 4)
	bc.addBlock(block2)
	fmt.Println("\n=== Blockchain after adding second block ===")
	bc.printChain()

	// 测试区块篡改（尝试修改第一个区块）
	if len(bc.blocks) > 0 {
		bc.blocks[0].Transactions[0].Amount = 1000 // 篡改金额
		fmt.Println("\n=== After tampering with genesis block ===")
		// 验证哈希是否失效
		tamperedHash := calculateHash(*bc.blocks[0])
		fmt.Printf("Original hash: %s\n", bc.blocks[0].Hash)
		fmt.Printf("New hash after tampering: %s\n", tamperedHash)
		fmt.Println("Tampering detected! Hash mismatch.")
	}
}
```
一、基础语法
 
- 变量与常量：
​
- 变量声明： var name type = value  或简写  name := value （类型推断）。
​
- 常量声明： const name type = value ，支持枚举式声明。
​
- 数据类型：
​
- 基本类型： int （随系统位数）、 float64 、 bool 、 string  等。
​
- 复合类型：数组（固定长度）、切片（动态长度， []type ）、映射（ map[keyType]valueType ）、结构体（ struct ）。
​
- 控制流：
​
- 条件： if 、 else if 、 else （条件无需括号）。
​
- 循环：仅  for  循环，支持类似  while  的用法（ for condition ）和无限循环（ for {} ）。
​
- 跳转： break 、 continue 、 goto （慎用）。
 
二、函数与方法
 
- 函数：
​
- 定义： func 函数名(参数列表) 返回值列表 { ... } ，支持多返回值（如  func add(a, b int) (int, error) ）。
​
- 匿名函数：可直接赋值给变量或立即执行（ func() { ... }() ）。
​
- 可变参数： func sum(nums ...int) int { ... } （参数以切片形式接收）。
​
- 方法：
​
- 与函数类似，但绑定到结构体，格式： func (接收者) 方法名(参数) 返回值 { ... } 。
​
- 接收者分值类型（复制结构体）和指针类型（修改原结构体）。
 
三、面向对象特性
 
- 结构体（struct）：替代类，用于封装数据。
​
- 接口（interface）：
​
- 定义： type 接口名 interface { 方法签名列表 } ，隐式实现（无需  implements  关键字）。
​
- 空接口  interface{}  可接收任意类型。
​
- 继承与组合：通过结构体嵌套实现组合（类似继承）。
 
四、并发编程
 
- goroutine：轻量级线程，由Go runtime管理，通过  go 函数名()  启动。
​
- 通道（channel）：
​
- 用于goroutine间通信， ch := make(chan 类型, 缓冲区大小) 。
​
- 操作： ch <- 数据 （发送）、 data := <-ch （接收），无缓冲通道会阻塞直到收发匹配。
​
- 同步机制：
​
-  sync.WaitGroup ：等待一组goroutine完成。
​
-  sync.Mutex / sync.RWMutex ：互斥锁，保证临界区安全。
​
-  select ：监听多个通道操作，类似  switch 。
 
五、包与模块
 
- 包（package）：
​
- 代码组织单位，文件名与包名可不同，入口函数为  main  包中的  main() 。
​
- 导出标识符：首字母大写（如  FuncName ）可被其他包访问。
​
- 模块（module）：
​
- Go 1.11+ 引入，通过  go mod init 模块名  初始化，管理依赖（ go get  安装依赖）。
 
六、错误处理与异常
 
- 错误处理：通过返回  error  类型处理预期错误（ if err != nil { ... } ）。
​
- 异常： panic  抛出异常（如数组越界）， recover  在  defer  中捕获异常，避免程序崩溃。
 
七、常用标准库
 
-  fmt ：输入输出。
​
-  os / os/exec ：系统操作、执行命令。
​
-  net/http ：HTTP服务与客户端（快速搭建Web服务）。
​
-  encoding/json ：JSON编解码。
​
-  sync ：并发同步工具。
​
-  time ：时间处理。
 
八、其他重要概念
 
- 切片（slice）：动态数组，由指针、长度、容量组成， append()  函数扩展长度。
​
- 映射（map）：无序键值对，必须初始化（ make(map[key]value) ）才能使用。
​
- 延迟执行（defer）： defer 语句  在函数返回前执行，常用于释放资源（如关闭文件）。
​
- 指针： *type  声明指针， &  取地址，不支持指针运算。