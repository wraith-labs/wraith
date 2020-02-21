# wraith-RAT

## Info

A Remote Administration Tool (RAT) written in Python with 
PHP/HTML/JS/CSS Command and Control (C&amp;C) API and panel.

## Installation Instructions (Latest - v3.0.0)

1) Download or clone this repository.

**SERVER**

2) Place the files in the `server` folder in the root of your HTTP server (Apache2 / PHP7 recommended).
3) Make sure that the required PHP extensions are installed (can be found in `info/required_libs.txt`).
4) If not using Apache2, make sure that the `server/assets/db.json`, `server/assets/wraith-scripts` files and directories are protected from public access (**IMPORTANT**). If using Apache2, this is already done using the `.htaccess` files.
5) If using Apache2, make sure `.htaccess` override is enabled in your Apache config.
6) Log into the panel by accessing the URL of your site (you should be automatically redirected to the login page). This is very important as it resets the encryption keys so that no one can access the API without logging in.
7) Change the panel login credentials in the `server/assets/db.json` file along with the wraith encryption key and the server fingerprint (any random strings, around 10-30 chars). Again, **VERY IMPORTANT**.

**CLIENT**

8) Make sure you are using `Python3.5` or above and have the libraries from `info/required_libs.txt` installed.
9) Go to a text hosting website such as `Pastebin` (from now on, intructions will refer to Pastebin) and make an account. You'll need it in order to later edit the file in case the address of your server changes.
10) Set the paste to never expire and set it's privacy to unlisted (optional but highly recommended)
11) Paste in the full address of your control server's API as the content; for example, `http://example.com/api.php`.
12) Edit the `client/wraith.py` file and change the contants to reflect your previously chosen settings. Should be self explanatory. (Warning: make sure the `FETCH_SERVER_LOCATION_URL` is a raw text URL; in other words, it has `/raw/` following `pastebin.com`)
13) Run the wraith in debug mode (defined by a constant in the file) first to verify that everything went well and the wraith is connecting to the server properly.
14) Log into the server to verify that commands are working. Try `ping` as the command to test if everything works.
15) Run the wraith without debug mode and enjoy. You can also freeze it with `PyInstaller` or others but only `PyInstaller` is officially supported.

NOTE: These installations only come with 2 basic payloads. For more pre-made payloads please see https://github.com/TR-SLimey/wraith-RAT-payloads

## Releases:
**v3.0.0:**
- First public release of wraith
- Basic functionality including:
  - Wraith successfully connects to the server
  - Wraith sends regular heartbeats to fetch commands and show signs of life
  - Wraith executes modular commands in threads
  - Wraith sends command results to the server
  - Server can manage multiple wraiths
  - Server can send modular commands
  - Server can receive command results
