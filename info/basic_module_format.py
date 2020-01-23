# This is an example of how modules should be formatted. It has no actual function

# args is a list - predefined by wraith
# wraith is an object - pre-defined by wraith
# cmd is a string - pre-defined by wraith

# Check if the wraith object is defined. If not, we are not running in the wraith
if not "wraith" in globals():
    print("This is a Wraith module and cannot run independently!")
    exit()

# Define a variable for the return value
return_val = None

# Begin the module
# Import required libraries. They must be in the wraith stdlib however.
import sys

try:
    # MODULE CODE HERE
    pass
except Exception as e:
    # In case of error, set return value to an error message
    return_val = "Error while running command {}. Err: {}".format(cmd, e)

# End by sending the wraith the return value
wraith.send(return_val)

# Always end with sys.exit(0) as each module will run in its own process
sys.exit(0)
