# wraith

## WARNING
The wraith client is in the process of being completely re-written (again, I know) for version 4.0.0. This update will slightly change the protocol but will mostly focus on shifting from `Python3` to `Golang` for the client, and will include an almost complete re-write of the control panel UI to make it much cleaner and more useable. Some "behind the scenes" changes on the panel side will also be made. A shift from a .JSON database file to a SQLite database is also planned on the server for better stability and maintainability. Here is a full overview of the planned changes:

 - Complete repository re-structure
 - Complete re-write of the panel UI to make it much cleaner and more useable
 - Some protocol changes to make Wraith communication more secure and robust
 - Complete re-write of the Wraith using Golang instead of Python3 mainly for speed, more low-level access, smaller filesize and compilation (as opposed to packaging with PyInstaller)
 - Re-designed modular payload delivery system to better work with Golang
 - Some panel "behind the scenes" changes such as moving to a more robust SQLite database
 - Settings to be moved into the panel over manually editing a file
 - More secure panel login system and ability to host the API and panel separately
 
The update does not yet have a planned release date. It will be released when all the abovementioned features are complete to a good standard and Wraith is suitably bug-free.

The update will also not be backwards-compatible with earlier Wraith or Panel versions due to the language transition and protocol changes but panel releases following this update are planned to be mostly backwards-compatible.

Previously, the Wraith project was intending to switch to C++ rather than Golang. However, Golang, while [slightly slower](https://benchmarksgame-team.pages.debian.net/benchmarksgame/fastest/go-gpp.html) than C++ (still faster than Python), is much more portable and easier for me to maintain.

To see the progress on the Wraith v4 release, please check the [indev](https://github.com/TR-SLimey/wraith-RAT/tree/indev) branch, though any code there is untested and not recommended for use.

Contributions to the Wraith project are very welcome. If you have a contribution you would like to make, feel free to open a pull request.

## Disclaimer
We, the project author and its contributors, do not accept any responsibility for any damage caused using this software. By using Wraith, you accept full responsibility for your actions and acknowledge that this software was created for educational purposes only, and meant for use exclusively on systems owned by the user, or those which the user has explicit permission to access for use with this software.

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
6) Log into the panel by accessing the URL of your site (you should be automatically redirected to the login page). This is very important as it resets the encryption keys so that no one can access the API without logging in. The credentials can be found in the `server/assets/db.json` file.
7) Change the panel login credentials in the `server/assets/db.json` file along with the wraith encryption key and the server fingerprint (any random strings, around 10-30 chars). Again, **VERY IMPORTANT**.

**CLIENT**

8) Make sure you are using `Python3.5` or above and have the libraries from `info/required_libs.txt` installed.
9) Go to a text hosting website such as `pastebin.com` (from now on, intructions will refer to Pastebin) and make an account. You'll need it in order to later edit the file in case the address of your server changes.
10) Set the paste to never expire and set it's privacy to unlisted (optional but highly recommended)
11) Paste in the full address of your control server's API as the content; for example, `http://example.com/api.php`.
12) Edit the `client/wraith.py` file and change the constants at the top of the file to reflect your previously chosen settings. Should be self explanatory. (Warning: make sure the `FETCH_SERVER_LOCATION_URL` is a raw text URL; in other words, it has `/raw/` following `pastebin.com`)
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
