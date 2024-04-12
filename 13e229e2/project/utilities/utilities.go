package utilities

import (
	"Driver-go/elevio"
	"math"
	"time"
)


func Delay_ms(duration int) {
	delay := time.NewTimer(time.Duration(duration) * time.Millisecond)
	<-delay.C
}
func Find_smallest(arr []elevio.ButtonEvent, lower int) int {
	smallest := arr[0].Floor
	for _, btn := range arr[1:] {
		if (btn.Floor < smallest && btn.Floor >= lower) && btn.Button == elevio.BT_HallUp {
			smallest = btn.Floor
		}
	}
	return smallest
}

func Find_largest(arr []elevio.ButtonEvent, upper int) int {
	largest := arr[0].Floor
	for _, btn := range arr[1:] {
		if (btn.Floor > largest && btn.Floor <= upper) && btn.Button == elevio.BT_HallDown {
			largest = btn.Floor
		}
	}
	return largest
}

func Abs_diff(a, b int) int {
	return int(math.Abs(float64(a - b)))
}

func Find_Index_of_min(arr []int) int {
	minIndex := 0
	minValue := arr[0]
	for index, value := range arr {
		if value < minValue {
			minIndex = index
			minValue = value
		}
	}
	return minIndex
}
