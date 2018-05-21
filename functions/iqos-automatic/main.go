package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sclevine/agouti"
	"os"
	"time"
)

var IQOS_PHERE_BASE_URL = "https://iqosphere.jp/"
var LOGIN_ID_ENV = "IQOS_LOGIN_ID"
var PASSWORD_ENV = "IQOS_PASSWORD"
var DEFAULT_LOGIN_ID = "xxxxxxxx"
var DEFAULT_PASSWORD = "xxxxxxxx"

type LambdaResponse struct {
	Message string
	Ok      bool
}

type LambdaEvent struct {
	LoginId  string
	Password string
}

func Handler(event LambdaEvent) (LambdaResponse, error) {
	fmt.Println("===== iQOS Crawling Start ==========")

	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			"--window-size=1920,1920",
		}), agouti.Debug,
	)
	err := driver.Start()
	chkErrAndExit(err)

	defer driver.Stop()
	page, err := driver.NewPage(agouti.Browser("chrome"))
	chkErrAndExit(err)

	loginIdText := event.LoginId
	if loginIdText == "" {
		loginIdText = os.Getenv(LOGIN_ID_ENV)
	}
	passwordText := event.Password
	if passwordText == "" {
		passwordText = os.Getenv(PASSWORD_ENV)
	}
	if loginIdText == "" || passwordText == "" {
		fmt.Println("ERROR: Incorrect user id and password.")
		os.Exit(1)
	}

	// Login page
	page.Navigate(IQOS_PHERE_BASE_URL + "login")
	sleep(3)
	fmt.Println("===> ✅Completed transition login page")
	identity := page.FindByName("login_id")
	password := page.FindByName("password")
	identity.Fill(loginIdText)
	password.Fill(passwordText)

	loginButton := page.FindByButton("ログイン")
	loginButton.Submit()

	sleep(5)
	fmt.Println("===> ✅Completed login")

	// Share page
	page.Navigate(IQOS_PHERE_BASE_URL + "miqos/share")
	sleep(5)
	fmt.Println("===> ✅Completed transition share page")

	sleep(5)

	var loop = 10
	likeTags := page.All("p.like")
	for i := 0; i < loop; i++ {
		likeTags.At(i).Click()
		fmt.Printf("===> ✅Clicked like(%d/%d)\n", i+1, loop)
		sleep(5)

		// Debug
		page.Screenshot(fmt.Sprintf("ss%d.png", i))
		//fmt.Printf("===> ✅Refresh page(%d/%d)\n", i+1, loop)
	}

	return LambdaResponse{
		Message: "success",
		Ok:      true,
	}, nil
}

func chkErrAndExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return
}

func sleep(sec int) {
	fmt.Printf("Wait %d sec...\n", sec)
	time.Sleep(time.Duration(sec) * time.Second)
}

func main() {
	debug := flag.Bool("debug", false, "Debug Mode(go run main.go -debug)")
	loginId := flag.String("u", "", "iQOS user id")
	passwd := flag.String("p", "", "iQOS password")
	flag.Parse()
	if *debug {
		fmt.Println("Run: Local Debug Mode")
		Handler(LambdaEvent{
			LoginId:  *loginId,
			Password: *passwd,
		})
	} else {
		fmt.Println("Run: Lambda Mode")
		lambda.Start(Handler)
	}
}
