# This command runs its arguments in the shell of the host os

def script_main(wraith, cmdname):
	try:
		sh_to_run = " ".join(cmdname.split(" ")[1:])
		wraith.putresult("SUCCESS", subprocess.run(sh_to_run, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=5, shell=True).stdout.decode('utf-8'))
	except Exception as e:
		wraith.putresult("ERROR - {}".format(e), "Error while executing `{}`".format(cmdname)) 
