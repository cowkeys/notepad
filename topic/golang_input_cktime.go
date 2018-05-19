package main

import (
	"flag"
	"fmt"
	"time"
)

func main() {
	local := flag.String("local", "Asia/Shanghai", "set location")
	stamp := flag.Int("s", 0, "timestamp to date")
	date := flag.String("d", "", "date to timestamp")
	flag.Parse()

	l, err := time.LoadLocation(*local)
	if err != nil {
		fmt.Println("err:", err)
		return
	}

	if *stamp > 0 {
		fmt.Printf("%d -> %v\n", *stamp, TimeFromStamp(int64(*stamp)))
	}

	if *date == "" {
		fmt.Printf("%v -> %d\n", time.Now(), time.Now().Unix())
		return
	}

	t, err := time.ParseInLocation("2006-01-02 15:04:05", *date, l)
	if err != nil {
		fmt.Printf("error:%v", err)
	}

	fmt.Printf("%v -> %d\n", t, t.Unix())
}

func TimeFromStamp(t int64) time.Time {
	if t == 0 {
		return time.Time{}
	}
	time := time.Unix(t, 0)
	return time
}

func TimeToStamp(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func TimeToMillisecond(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano() / int64(time.Millisecond)
}
