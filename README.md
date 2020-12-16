# What is this?
This repository can serve as a example for your own code, or can be used as a full solution.  
The program will log into your account on epicgames.com and add all free games that it finds to your account.  
It uses hCaptcha's own hCaptcha bypass mechanism B-).  
No data is shared with third parties or me.  
I also am not responsible for any misery caused to you by this program.  
# Getting Started
There are two methods of automatically adding all of these games, manual and automatic.
## Manual
- Fetch the proper [release](https://github.com/hb0nes/epic-store-free-games-snatcher/releases). 
- Go [here](https://dashboard.hcaptcha.com/signup?type=accessibility) and get an accessibility URL by filling in your e-mail. The mail that they send you contains a button.  
Right click this button and get the URL, and insert that URL in config.yaml.example    
This example file also shows what it should look like. If you want to use one key, remove the example so that you only have one entry left (your own key).  
- Fill in your username & password as well, to be able to log into the Epic Games store.  
- Rename config.yaml.example to config.yaml and make sure it is in the same directory as your executable.
- Make sure you have Chrome installed, and run the program!  
IMPORTANT - Epic Games Store only accepts e-mail, not your actual username, apparently.
## Automatic
`Regarding the secrets, follow the instructions above.`  
There is no support for 2FA (yet).  
- [Fork](https://github.com/hb0nes/epic-store-free-games-snatcher/fork)  this repository
- Add three GitHub secrets:
  - hCaptchaURL
  - username
  - password
- Enable [Git Actions](https://docs.github.com/en/free-pro-team@latest/actions/managing-workflow-runs/disabling-and-enabling-a-workflow) and enable the 'Go' workflow.
- It will run twice per day, each day. Do not adjust the interval, hCaptcha has very strict ratelimits.  


If it all fails, open an issue and I might look into it.
I might also not.