# This is an example of how modules should be formatted. It has no actual function

# SOME INFO

# Pre-defined variables which can be used:
#	args is a list - predefined by wraith - arguments passed to the command
#	cmd is a string - predefined by wraith - the name of the command
#	wraith is an object - predefined by wraith - the wraith object. Be careful!

# These modules run in threads started by the wraith and by default, have access to every
# library used by the wraith. Additional libraries can be imported but this is not recommended
# as the commands will then be unable to run if the wraith is frozen (PyInstaller, etc.)

# Warning: Comments are not stripped by the panel (TODO) so the less comments the better as
# less data needs to be sent.

# END OF INFO

# Check if the wraith object is defined. If not, we are not running in the wraith
if not "wraith" in globals():
    print("This is a Wraith module and cannot run independently!")
    exit()

# Define a variable for the return value
return_val = None

try:
    # MODULE CODE HERE
    pass
except Exception as e:
    # In case of error, set return value to an error message
    return_val = "Error while running command {}. Err: {}".format(cmd, e)

# End by sending the wraith the return value
wraith.returncmd(return_val)

