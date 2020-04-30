package main

/*

This file contains the constants used at compile-time. These will be hard-coded
into the Wraith when it is compiled. Optionally, these can be overridden using
command line parameters on run if this is enabled below.

*/

/* Link to *raw text* page containing the URL of the C&C server. Pastebin
is a good choice but any raw text hosting will do.
<STRING> */
const setCCSERVERGETURL string = "http://localhost/location.txt" //"https://pastebin.com/raw/snBS6LEk" //"http://pastebin.com/raw/56BqPLEJ"

/* */
const setAPIREQPREFIX string = "W_"

/* Static encryption key to use during the handshake with the server. It will
be replaced with a panel-generated key after the handshake until the Wraith
is restarted or disconnects. This key does not have to be very secure,
the primary encryption method is SSL. This is really only used to make Wraith's
traffic less readable to decrypting proxies.
<STRING> */
const setSECONDLAYERENCRYPTIONKEY string = "QWERTYUIOPASDFGHJKLZXCVBNM"

/* An ID which the server uses to identify itself. The Wraith will not complete
a handshake with a server which does not provide a trusted fingerprint. This is
used to stop accidental connections to the wrong server if the URL is wrong and
prevent server impersonation.
<STRING> */
const setTRUSTEDSERVERFINGERPRINT string = "ABCDEFGHIJKLMNOP"

/* */
const setINSTALLONRUN bool = false

/* Delay between each heartbeat request (seconds). This should ideally be just
under half of the mark dead time on the panel and no less than 3 seconds to
prevent DDoSing your own server. This value is not permanent and can be
altered by the panel during runtime but will reset whenever the Wraith
restarts.
<UNSIGNED LONG> */
const setDEFAULTHEARTBEATDELAYBASE uint64 = 10

/* Delay between handshake reattempts should a handshake fail (seconds). This
should be longer than the heartbeat delay as the server might be unreachable
for a while so the less traffic we send the better.
<UNSIGNED LONG> */
const setDEFAULTHANDSHAKEREATTEMPTDELAY uint64 = 10 // 20

/* */
const setFAILEDHEARTBEATTOLERANCE uint64 = 2

/* How long to wait for the response to each HTTP request before cancelling
the request and producing a (caught) error.
<UNSIGNED LONG> */
const setDEFAULTHTTPREQUESTTIMEOUT uint64 = 5

/* Whether to write information to the console. It is highly recommended to
turn this off as leaving it on increases the executable filesize and could
leak information.
<NO VALUE> */
const setDEBUG bool = true
