package main

import (
	"fmt"
	"os"
	"strconv"
)

func printAll() {
	fmt.Println("Calculator")
	fmt.Println("一个实现加减乘除取余的计算器")
	fmt.Println("使用方法：calculator <操作符> <数字1> <数字2>")
	fmt.Println("+ 加号")
	fmt.Println("- 减号")
	fmt.Println("* 乘号")
	fmt.Println("/ 除号")
	fmt.Println("% 取余号")
	fmt.Println("示例：calculator + 3 4")
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("参数量不匹配")
		printAll()
		os.Exit(1)
	}

	operator := os.Args[1]
	num1 := os.Args[2]
	num2 := os.Args[3]
	num1Trans, err1 := strconv.ParseFloat(num1, 64)
	num2Trans, err2 := strconv.ParseFloat(num2, 64)
	if err1 != nil || err2 != nil {
		fmt.Println("输入的数字有误")
		printAll()
		os.Exit(1)
	}

	var result float64
	var err error

	switch operator {
	case "+":
		result = num1Trans + num2Trans
	case "-":
		result = num1Trans - num2Trans
	case "*":
		result = num1Trans * num2Trans
	case "/":
		if num2Trans == 0 {
			err = fmt.Errorf("除数不能为零")
		} else {
			result = num1Trans / num2Trans
		}
	case "%":
		if num2Trans == 0 {
			err = fmt.Errorf("除数不能是0")
		} else {
			// 检查是否为整数
			if num1Trans != float64(int(num1Trans)) || num2Trans != float64(int(num2Trans)) {
				err = fmt.Errorf("取余运算只支持整数")
			} else {
				result = float64(int(num1Trans) % int(num2Trans))
			}
		}
	default:
		err = fmt.Errorf("不支持的操作符: %s", operator)
	}

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		printAll()
		os.Exit(1)
	}

	// 输出结果，根据情况显示为整数或浮点数
	if result == float64(int(result)) {
		fmt.Printf("%d\n", int(result))
	} else {
		fmt.Printf("%g\n", result)
	}
}
