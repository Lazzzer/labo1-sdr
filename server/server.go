package main

import (
	"fmt"
	"github.com/Lazzzer/labo1-sdr/utils/types"
)

func main() {
	job := types.Job{
		Id:           "1",
		Name:         "Job 1",
		NbVolunteers: 200,
	}
	fmt.Println(job)
}
