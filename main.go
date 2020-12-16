package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

var config Config

func doSlowCall(c context.Context, a chromedp.QueryAction, seconds time.Duration) error {
	ctx, cancel := context.WithTimeout(c, time.Second*seconds)
	defer cancel()
	err := a.Do(ctx)
	return err
}

// func handleFreeGames(c context.Context, nodes []*cdp.Node) {
func handleFreeGames(c context.Context, urls []string) {
	fmt.Printf("Handling %d games!\n", len(urls))
	for _, url := range urls {
		fmt.Printf("Checking URL: %s\n", url)
		chromedp.Navigate(url).Do(c)
		if err := doSlowCall(c, chromedp.WaitEnabled(`//button[text()[contains(.,"Continue")]]`), 5); err == nil {
			doSlowCall(c, chromedp.Click(`//button[text()[contains(.,"Continue")]]`), 5)
		}
		fmt.Println("Checking if already owned.")
		if err := doSlowCall(c, chromedp.WaitVisible(`//button[./span[text()[contains(.,"Owned")]]]`), 1); err == nil {
			fmt.Println("Already owned. Skipping.")
			continue
		}
		fmt.Println("Waiting for GET button")
		if err := doSlowCall(c, chromedp.WaitEnabled(`//button[@data-testid="purchase-cta-button"]`), 5); err == nil {
			fmt.Println("Clicking GET button")
			chromedp.Sleep(time.Second).Do(c)
			chromedp.Click(`//button[@data-testid="purchase-cta-button"]`).Do(c)
		} else {
			fmt.Println("Could not find the GET button. Skipping.")
			continue
		}
		fmt.Println("Waiting for checkbox")
		if err := doSlowCall(c, chromedp.WaitEnabled(`//i[@class="icon-checkbox-unchecked radio-unchecked"]`), 5); err == nil {
			fmt.Println("Clicking checkbox")
			chromedp.Sleep(time.Second).Do(c)
			chromedp.Click(`//i[@class="icon-checkbox-unchecked radio-unchecked"]`).Do(c)
		} else {
			fmt.Println("Could not find the Checkbox. Skipping.")
			continue
		}
		fmt.Println("Waiting for Place Order")
		if err := doSlowCall(c, chromedp.WaitEnabled(`//button[./span[text()[contains(.,"Place Order")]]]`), 5); err == nil {
			fmt.Println("Clicking Place Order")
			chromedp.Sleep(time.Second).Do(c)
			chromedp.Click(`//button[./span[text()[contains(.,"Place Order")]]]`).Do(c)
		} else {
			fmt.Println("Could not find the Place Order button. Skipping.")
			continue
		}
		fmt.Println("Waiting for Agreement")
		if err := doSlowCall(c, chromedp.WaitEnabled(`//button[span[text()="I Agree"]]`), 5); err == nil {
			fmt.Println("Clicking I Agree")
			chromedp.Sleep(time.Second).Do(c)
			chromedp.Click(`//button[span[text()="I Agree"]]`).Do(c)
		} else {
			fmt.Println("Could not find the 'I Agree' button. Skipping.")
			continue
		}
		doSlowCall(c, chromedp.WaitEnabled(`//span[text()="Thank you for buying"]`), 5)
	}
}

func getFreeGameURLs(ctx context.Context) (urls []string) {
	chromedp.Run(ctx,
		chromedp.Navigate("https://www.epicgames.com/store/en-US/free-games"),
		chromedp.WaitVisible(`//a[@aria-label[contains(., "Free Games")]]`),
		chromedp.Sleep(time.Second*5),
		chromedp.QueryAfter(`//a[@aria-label[contains(., "Free Games")]]`, func(ctx context.Context, bla runtime.ExecutionContextID, nodes ...*cdp.Node) error {
			if len(nodes) < 1 {
				return errors.New("expected at least one node")
			}
			for _, node := range nodes {
				url, _ := node.Attribute("href")
				urls = append(urls, "https://www.epicgames.com"+url)
			}
			return nil
		}))
	return urls
}

func getAccessibilityCookie(ctx context.Context) {
	tries := 5
	for _, link := range config.HCaptchaURLs {
		chromedp.Navigate(link).Do(ctx)
		chromedp.WaitEnabled(`//button[@data-cy='setAccessibilityCookie']`).Do(ctx)
		for i := 0; i < tries; i++ {
			fmt.Printf("Trying to get hCaptcha accessibility cookie... %d of %d\n", i+1, tries)
			chromedp.Sleep(time.Millisecond * time.Duration(rand.Intn(2500)+2500)).Do(ctx)
			chromedp.Click(`//button[@data-cy='setAccessibilityCookie']`).Do(ctx)
			chromedp.Sleep(time.Millisecond * time.Duration(rand.Intn(2500)+2500)).Do(ctx)
			if accessibilityCookie, _ := checkCookies(ctx); accessibilityCookie {
				fmt.Println("Acquired hCaptcha cookie!")
				return
			}
		}
	}
	log.Fatal("Couldn't get accessibility cookie from hCaptcha, so cannot bypass captcha. Try again another time.")
}

func getEpicStoreCookie(ctx context.Context) {
	fmt.Println("Logging into Epic Games Store.")
	tries := 5
	chromedp.Run(ctx,
		chromedp.Navigate(`https://www.epicgames.com/login`),
		chromedp.WaitEnabled(`//div[@id='login-with-epic']`),
		chromedp.Click(`//div[@id='login-with-epic']`),
		chromedp.WaitEnabled(`//input[@id='email']`),
		chromedp.SendKeys(`//input[@id='email']`, config.Username),
		chromedp.WaitEnabled(`//input[@id='password']`),
		chromedp.SendKeys(`//input[@id='password']`, config.Password),
	)
	for i := 0; i < tries; i++ {
		fmt.Printf("Trying to log in to Epic Games Store... %d of %d\n", i+1, tries)
		if err := doSlowCall(ctx, chromedp.WaitEnabled(`//button[@id='sign-in']`), 5); err == nil {
			doSlowCall(ctx, chromedp.Click(`//button[@id='sign-in']`), 2)
		}
		chromedp.Sleep(time.Second).Do(ctx)
		if _, epicStoreCookie := checkCookies(ctx); epicStoreCookie {
			fmt.Println("Logged into Epic Games Store.")
			return
		}
	}
	log.Fatal("Apparently, logging in is not successful. Too bad.")
}

func checkCookies(ctx context.Context) (accessibilityCookie bool, epicCookie bool) {
	cookies, err := network.GetAllCookies().Do(ctx)
	if err != nil {
		panic(err)
	}
	for _, cookie := range cookies {
		if cookie.Name == "hc_accessibility" {
			accessibilityCookie = true
		}
		if cookie.Name == "EPIC_SSO" {
			epicCookie = true
		}
	}
	return accessibilityCookie, epicCookie
}

func getCookies(ctx context.Context) {
	accessibilityCookie, epicGamesCookie := checkCookies(ctx)
	if epicGamesCookie {
		fmt.Println("Existing cookie found for Epic Games Store. Doing nothing =].")
		return
	}
	if !accessibilityCookie {
		fmt.Println("Need to bypass hCaptcha to login to Epic Games Store.")
		getAccessibilityCookie(ctx)
	}
	getEpicStoreCookie(ctx)
}

func setupLogger(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			if e, ok := ev.(*runtime.EventConsoleAPICalled); ok {
				for _, arg := range e.Args {
					if e.Type == runtime.APITypeError && arg.Type != runtime.TypeUndefined {
						fmt.Printf("%s\n", arg.Value)
					}
				}
			}
		}()
	})
}

func setCookies(ctx context.Context) {
	fmt.Println("Setting cookies.")
	expiryTime := cdp.TimeSinceEpoch(time.Now().Add(time.Hour))
	cookies := []*network.CookieParam{
		{
			Name:    "OptanonAlertBoxClosed",
			Value:   "en-US",
			URL:     "epicgames.com",
			Expires: &expiryTime,
		},
	}
	network.SetCookies(cookies).Do(ctx)
}

func main() {
	config = readConfig()
	dir, err := ioutil.TempDir("", "free-game-fetcher-2000")
	if err != nil {
		log.Fatalf("Could not create user data dir for Chrome in %s\n", dir)
	}
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.UserDataDir(dir),
		chromedp.DisableGPU,
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("start-maximized", true),
	}
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	if err := chromedp.Run(taskCtx); err != nil {
		log.Fatalf("Could not start Chrome?\n%s\n", err)
	}
	if err := chromedp.Run(taskCtx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			setupLogger(ctx)
			getCookies(ctx)
			setCookies(ctx)
			handleFreeGames(ctx, getFreeGameURLs(ctx))
			fmt.Println("Done!")
			return nil
		}),
	); err != nil {
		log.Fatal(err)
	}
}
