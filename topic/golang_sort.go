package main

import (
	"fmt"
)

var (
	origin = []int{10, 25, 2, 8, 1, 13, 54, 101, 9, 17}
)

func main() {
	//pop()
	//selectsort()
	insertsort()
}

//冒泡
func pop() {
	fmt.Printf("origin = %v \n", origin)
	length := len(origin)
	for i := 0; i < length; i++ {
		for j := 0; j < length-1; j++ {
			if origin[j] <= origin[j+1] {
				continue
			}

			temp := origin[j]
			origin[j] = origin[j+1]
			origin[j+1] = temp
		}
	}

	fmt.Printf("sorted = %v", origin)
}

//选择
func selectsort() {
	fmt.Printf("origin = %v \n", origin)
	length := len(origin)
	for i := 0; i < length-1; i++ {
		min := i
		for j := i + 1; j < length; j++ {
			if origin[j] < origin[min] {
				min = j
			}
		}

		temp := origin[i]
		origin[i] = origin[min]
		origin[min] = temp
	}
	fmt.Printf("sorted = %v", origin)
}

//插入排序
func insertsort() {
	fmt.Printf("origin = %v \n", origin)

	length := len(origin)

	for i := 0; i < length; i++ {
		min := 0
		for j := i; j > 0; j-- {
			if origin[j] < origin[i] {
				min = j
			}
		}

		temp := origin[i]
		origin[i] = origin[min]
		origin[min] = temp
	}

	fmt.Printf("sorted = %v", origin)
}
