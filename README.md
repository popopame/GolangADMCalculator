# Golang ADM Calculator

I created this script to automate the calculation of the Accelerated Dual Momentum.
Once the script computed the Result, a mail is sent to a defined list of email.

I didn't create the ADM approach, please read this blog post to know every technical aspect of this asset allocation approach: https://engineeredportfolio.com/2018/05/02/accelerating-dual-momentum-investing/.

Feel free to Modify the code if you see any possible Optimization.

## What you will need to use this script

**A Yahoo Finance API Account**
This script use information requested on the Yahoo Finance API, an account must be created on the Yahoo Finance API website: https://www.yahoofinanceapi.com
For this script, the free plan is more than enough.
Once the account created, retrieve the API Key on the Dashboard:https://www.yahoofinanceapi.com/dashboard

**A  SMTP Server with an Email Address**
The result of the ADM are sent via mail, and so, a working and accessible SMTP server is needed, with an email configured on it.

## How to Configure and use the script

Most of the information needed by the script will be configured on the `config.yaml` file.
Every argument is a string.

But two variable can be passed via Environement Variable or Flag:
* *Yahoo API Key*: Can be set via the `key` flag, the `YAHOO_API_KEY` environment variable or the `apikey` arg in the yaml 
* *Email Address Passorwd*: Can be set via the `smtppasswd` flag, the `SMTP_PASSWD` environment variable, or the smtpPasswd arg in the yaml.

The rest of the Variable are set in the `config.yaml` file:
* *receivers*: List of email that will receive the mail, set as an array.
* *sender*: Mail used to send the mail
* *smtpAddr*: Address of the smtp server
* *smtpPort*: Port used by the smtp server

### Use the script From Command line

Make sure that you have go installed on your computer.
Populate the config file `config.yaml` and run the code.

Without building:
`go run ADMCalculator.go`

Or build and run the code:
`go build && ./ADMCalculator.go`

You can pass flag when calling the script, add the `-h` flag for a list of flag.