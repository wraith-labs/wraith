# This is an example of how modules should be formatted. It has no actual function

# SOME INFO

# All code should be put in the `script_main` function which is executed by the wraith
# in a separate thread. Any code outside of this function will be executed in the command
# executor thread which can delay or block other commands to be ran by the same thread.

# The `script_main` function will have 2 arguments passed to it; the wraith object and
# the full command line respectively. It is up to the module to process this command line
# such as by splitting it into separate arguments. The first word of the command line will
# always be the command itself. For example, "echo hello there" will be the command line
# of a command called "echo" with the arguments "hello there".

# These modules run in threads started by the wraith and by default, have access to every
# library used by the wraith. Additional libraries can be imported here but this is not
# recommended as the commands will then be unable to run if the wraith is frozen (PyInstaller, etc.)

# Warning: Comments are not stripped by the panel so the less comments the better as
# less data needs to be sent.

# END OF INFO

# The following is an example of the `sh` command:

def script_main(wraith, cmdname):
	try:
		sh_to_run = " ".join(cmdname.split(" ")[1:])
		wraith.putresult("SUCCESS", subprocess.run(sh_to_run, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, stdin=subprocess.DEVNULL, timeout=5, shell=True).stdout.decode('utf-8'))
	except Exception as e:
		wraith.putresult("ERROR - {}".format(e), "Error while executing `{}`".format(cmdname)) 
