package cronjobs

import (
	"fmt"
	"testing"
	"time"
)

type Jobs struct {
}

func (j Jobs) Run() {
	fmt.Println("job is running....")
}

func TestCronInstance(t *testing.T) {
	instance := CronInstance()
	t.Run("Test cron instance init", func(t *testing.T) {

		//if instance == nil {
		//	t.Error("Cron Instance init fail.")
		//}
		//instance1 := CronInstance()
		//if instance != instance1 {
		//	t.Error("Cron Instance is change.")
		//}

		jobs := Jobs{}
		//err := DoOnce(jobs)
		//t.Error(err)

		//CronInstance().AddJob("@every 3s", jobs)
		//
		//CronInstance().AddFunc("@every 1s", func() {
		//	//t.Log("func running in 1 sec")
		//	fmt.Println("func running in 1 sec....")
		//})

		DoOnce(jobs)

	})

	instance.Start()

	time.Sleep(time.Second * 10)
}