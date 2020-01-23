'use strict';

const crypto = require('crypto');
const fs = require('fs');


/**
 * Encrypts data and files using AES CBC/CFB - 128/192/256 bits. 
 * 
 * The encryption and authentication keys 
 * are derived from the supplied key/password using HKDF/PBKDF2.
 * The key can be set either with `setMasterKey` or with `randomKeyGen`.
 * Encrypted data format: salt[16] + iv[16] + ciphertext[n] + mac[32].
 * Ciphertext authenticity is verified with HMAC SHA256.
 * 
 * @property {Number} keyIterations The number of PBKDF2 iterations.
 * @property {Boolean} base64 Accepts ans returns base64 encoded data.
 */
class AesEncryption {
    /** Creates a new AesEncryption object.
     * @param {String} [mode=cbc] Optional, the AES mode (cbc or cfb)
     * @param {Number} [size=128] Optional, the key size (128, 192 or 256)
     * @throws {Error} if the mode is not supported or key size is invalid.
     */
    constructor(mode, size) {
        mode = (mode === undefined) ? 'cbc' : mode.toLowerCase();
        size = (size === undefined) ? 128 : size;

        if (!AES.Modes.hasOwnProperty(mode)) {
            throw Error(mode + ' is not supported!')
        }
        if (AES.Sizes.indexOf(size) == -1) {
            throw Error('Invalid key size!')
        }
        this._keyLen = size / 8;
        this._cipher = AES.Modes[mode].replace('size', size);
        this._masterKey = null;

        this.keyIterations = 20000;
        this.base64 = true;
    }

    /**
     * Encrypts data using a key or the supplied password.
     *
     * The password is not required if the master key has been set - 
     * either with `randomKeyGen` or with `setMasterKey`. 
     * If a password is supplied, it will be used to create a key with PBKDF2.
     * 
     * @param {(Buffer|String)} data The plaintext.
     * @param {String} [password=null] Optional, the password.
     * @return {(Buffer|String)} Encrypted data (salt + iv + ciphertext + mac).
     */
    encrypt(data, password) {
        const salt = randomBytes(saltLen);
        const iv = randomBytes(ivLen);
        try {
            const _keys = keys.call(this, salt, password);
            const aesKey = _keys[0], macKey = _keys[1];
            
            const aes = cipher.call(this, aesKey, iv, AES.Encrypt);
            const ciphertext = Buffer.concat(
                [iv, aes.update(data), aes.final()]
            );
            const mac = sign(ciphertext, macKey);
            let encrypted = Buffer.concat([salt, ciphertext, mac]);
            if (this.base64) {
                encrypted = encrypted.toString('base64');
            }
            return encrypted;
        } catch (err) {
            this._errorHandler(err);
            return null;
        }
    }

    /**
     * Decrypts data using a key or the supplied password.
     * 
     * The password is not required if the master key has been set - 
     * either with `randomKeyGen` or with `setMasterKey`. 
     * If a password is supplied, it will be used to create a key with PBKDF2.
     * 
     * @param {(Buffer|String)} data The ciphertext.
     * @param {String} [password=null] Optional, the password.
     * @return {(Buffer|String)} Plaintext.
     */
    decrypt(data, password) {
        try {
            if (this.base64) {
                data = Buffer.from(data, 'base64')
            }
            const salt = data.slice(0, saltLen);
            const iv = data.slice(saltLen, saltLen + ivLen);
            const ciphertext = data.slice(saltLen + ivLen, -macLen);
            const mac = data.slice(-macLen, data.length);

            const _keys = keys.call(this, salt, password);
            const aesKey = _keys[0], macKey = _keys[1];

            verify(Buffer.concat([iv, ciphertext]), mac, macKey);
            const aes = cipher.call(this, aesKey, iv, AES.Decrypt);
            const plaintext = Buffer.concat(
                [aes.update(ciphertext), aes.final()]
            );
            return plaintext;
        } catch (err) {
            this._errorHandler(err);
            return null;
        }
    }

    /**
     * Encrypts files using a master key or the supplied password.
     * 
     * The original file is not modified; a new encrypted file is created.
     * The password is not required if the master key has been set - 
     * either with `randomKeyGen` or with `setMasterKey`. 
     * If a password is supplied, it will be used to create a key with PBKDF2.
     * 
     * @param {String} path The file path.
     * @param {String} [password=null] Optional, the password.
     * @return {String} The new file path.
     */
    encryptFile(path, password) {
        const salt = randomBytes(saltLen);
        const iv = randomBytes(ivLen);
        try {
            const _keys = keys.call(this, salt, password);
            const aesKey = _keys[0], macKey = _keys[1];
            const aes = cipher.call(this, aesKey, iv, AES.Encrypt);
            const hmac = crypto.createHmac('sha256', macKey);

            const newPath = path + '.enc';
            const fd = fs.openSync(newPath, 'w');
            const chunks = fileChunks(path);

            fs.writeSync(fd, salt, 0, saltLen);
            fs.writeSync(fd, iv, 0, ivLen);
            hmac.update(iv);
            do {
                var chunk = chunks.next();
                var data = aes.update(chunk.value || '');

                fs.writeSync(fd, data, 0, data.length);
                hmac.update(data);
            } while (!chunk.done);
            
            data = aes.final();
            fs.writeSync(fd, data, 0, data.length);
            hmac.update(data);

            const mac = hmac.digest();
            fs.writeSync(fd, mac, 0, macLen);
            fs.closeSync(fd);
            return newPath;
        } catch (err) {
            this._errorHandler(err);
            return null;
        }
    }

    /**
     * Decrypts files using a master key or the supplied password.
     * 
     * The original file is not modified; a new decrypted file is created.
     * The password is not required if the master key has been set - 
     * either with `randomKeyGen` or with `setMasterKey`. 
     * If a password is supplied, it will be used to create a key with PBKDF2.
     * 
     * @param {String} path The file path. 
     * @param {String} [password=null] Optional, the password.
     * @return {String} The new file path.
     */
    decryptFile(path, password) {
        try {
            const salt = Buffer.alloc(saltLen);
            const iv = Buffer.alloc(ivLen);
            const mac = Buffer.alloc(macLen);

            const fileSize = fs.statSync(path).size;
            let fd = fs.openSync(path, 'r');
            fs.readSync(fd, salt, 0, saltLen);
            fs.readSync(fd, iv, 0, ivLen);
            fs.readSync(fd, mac, 0, macLen, fileSize - macLen);
            fs.closeSync(fd);

            const _keys = keys.call(this, salt, password);
            const aesKey = _keys[0], macKey = _keys[1];
            verifyFile(path, mac, macKey);

            const aes = cipher.call(this, aesKey, iv, AES.Decrypt);
            const newPath = path.replace(/\.enc$/, '.dec');
            fd = fs.openSync(newPath, 'w');
            const chunks = fileChunks(path, saltLen + ivLen, macLen);
            do {
                var chunk = chunks.next();
                var data = aes.update(chunk.value || '');
                fs.writeSync(fd, data, 0, data.length);
            } while (!chunk.done);
            
            data = aes.final();
            fs.writeSync(fd, data, 0, data.length);
            fs.closeSync(fd);
            return newPath;
        } catch (err) {
            this._errorHandler(err);
            return null;
        }
    }

    /**
     * Sets a new master key.
     * This key will be used to create the encryption and authentication keys.
     * 
     * @param {(Buffer|String)} key The new master key.
     * @param {Boolean} [raw=false] Optional, expexts raw bytes (not base64-encoded).
     */
    setMasterKey(key, raw) {
        try {
            key = (raw !== true) ? Buffer.from(key, 'base64') : key;
            if (!(key instanceof Buffer)) {
                throw Error('Key must be a Buffer!');
            }
            this._masterKey = key;
        } catch (err) {
            this._errorHandler(err);
        }
    }

    /**
     * Returns the master key (or null if the key is not set).
     * 
     * @param {Boolean} [raw=false] Optional, returns raw bytes (not base64-encoded).
     * @return {(Buffer|String)} The master key.
     */
    getMasterKey(raw) {
        if (this._masterKey === null) {
            this._errorHandler(new Error('The key is not set!'));
        } else if (raw !== true) {
            return this._masterKey.toString('base64');
        }
        return this._masterKey;
    }

    /**
     * Generates a new random key.
     * This key will be used to create the encryption and authentication keys.
     * 
     * @param {Number} [keyLen=32] Optional, the key size.
     * @param {Boolean} [raw=false] Optional, returns raw bytes (not base64-encoded).
     * @return {(Buffer|String)} The new master key.
     */
    randomKeyGen(keyLen, raw) {
        keyLen = (keyLen !== undefined) ? keyLen : 32;
        this._masterKey = randomBytes(keyLen);

        if (raw !== true) {
            return this._masterKey.toString('base64');
        }
        return this._masterKey;
    }

    /**
     * Handles exceptions (prints the error message by default).
     */
    _errorHandler(error) {
        console.log(error.message);
    }
}


module.exports = AesEncryption;


const saltLen = 16;
const ivLen = 16;
const macLen = 32;
const macKeyLen = 32;

const AES = {
    Modes: {'cbc': 'aes-size-cbc', 'cfb': 'aes-size-cfb8'}, 
    Sizes: [128, 192, 256], 
    Encrypt: 1,
    Decrypt: 2
};

/**
 * Creates random bytes, used for IV, salt and key generation.
 */
function randomBytes(size) {
    return crypto.randomBytes(size);
}

/**
 * Creates a crypto.cipher object, used for encryption.
 */
function cipher(key, iv, operation) {
    if (operation === AES.Encrypt) {
        return crypto.createCipheriv(this._cipher, key, iv);
    } else if (operation === AES.Decrypt) {
        return crypto.createDecipheriv(this._cipher, key, iv);
    } else {
        throw Error('Invalid operation!');
    }
}

/**
 * Derives encryption and authentication keys from a key or password.
 * If the password is not null, it will be used to create the keys.
 * 
 * @throws {Error} If neither the key or password is set.
 */
function keys(salt, password) {
    if (password !== undefined && password !== null) {
        var dkey = crypto.pbkdf2Sync(
            password, salt, this.keyIterations, this._keyLen + macKeyLen, 'sha512'
        );
    } else if (this._masterKey !== null) {
        var dkey = hkdfSha256(this._masterKey, this._keyLen + macKeyLen, salt)
    } else {
        throw Error('No password or key specified!');
    }
    return [
        dkey.slice(0, this._keyLen), 
        dkey.slice(this._keyLen, dkey.length)
    ]
}

/**
 * Computes the MAC of ciphertext, used for authentication.
 */
function sign(ciphertext, key) {
    const hmac = crypto.createHmac('sha256', key);
    hmac.update(ciphertext);
    return hmac.digest();
}

/**
 * Verifies the authenticity of ciphertext.
 * @throws {Error} if the MAC is invalid.
 */
function verify(ciphertext, mac, key) {
    const ciphertextMac = sign(ciphertext, key);
    if (!constantTimeComparison(mac, ciphertextMac)) {
        throw Error('Mac check failed!');
    }
}

/**
 * Computes the MAC of ciphertext, used for authentication.
 */
function signFile(path, key, fbeg, fend) {
    const hmac = crypto.createHmac('sha256', key);
    const chunks = fileChunks(path, fbeg, fend);
    do {
        var chunk = chunks.next();
        hmac.update(chunk.value || Buffer.alloc(0));
    } while (!chunk.done);
    return hmac.digest();
}

/**
 * Verifies the authenticity of ciphertext.
 * @throws {Error} if the MAC is invalid.
 */
function verifyFile(path, mac, key) {
    const fileMac = signFile(path, key, saltLen, macLen);
    if (!constantTimeComparison(mac, fileMac)) {
        throw Error('Mac check failed!');
    }
}

/**
 * Safely compares two byte arrays, used for ciphertext uthentication.
 */
function constantTimeComparison(macA, macB) {
    let result = macA.length ^ macB.length;
    for (let i=0; i<macA.length && i< macB.length; i++) {
        result |= macA[i] ^ macB[i];
    }
    return result === 0;
}

/**
 * A generator that reads a file and yields chunks of data.
 * 
 * @param {String} path The file path.
 * @param {Number} [beg=0] Optional, the start position.
 * @param {Number} [end=0] Optional, the end position.
 * @yield {Buffer} File data.
 */
function* fileChunks(path, beg, end) {
    beg = (beg === undefined) ? 0 : beg;
    end = fs.statSync(path).size - ((end === undefined) ? 0 : end);
    
    let size = 1024;
    const fp = fs.openSync(path, 'r');
    const buffer = Buffer.alloc(size);
    let pos = fs.readSync(fp, Buffer.alloc(beg + 1), 0, beg);

    while (pos < end) {
        size = (end - pos > size) ? size : (end - pos);
        let chunkSize = fs.readSync(fp, buffer, 0, size);
        pos += chunkSize;
        yield buffer.slice(0, chunkSize);
    }
}

/**
 * A HKDF implementation, with HMAC SHA256.
 * Used for expanding the master key and derive AES and HMAC keys.
 * 
 * @param {Buffer} key The master key.
 * @param {Number} keySize The size of the derived key.
 * @param {Buffer} [salt=null] Optional, the salt (random bytes).
 * @param {Buffer} [info=null] Optional, information about the key.
 * @return {Buffer} Derived key material.
 */
function hkdfSha256(key, keySize, salt, info) {
    let dkey = Buffer.alloc(0);
    let hmac = crypto.createHmac('sha256', salt || '');
    const prk = hmac.update(key).digest();
    const hashLen = 32;

    for (let i = 0; i < Math.ceil(1.0 * keySize / hashLen); i++) {
        hmac = crypto.createHmac('sha256', prk);
        hmac.update(Buffer.concat([
            dkey.slice(dkey.length - hashLen), 
            Buffer.from(info || ''), Buffer.alloc(1, i + 1)
        ]));
        dkey = Buffer.concat([dkey, hmac.digest()]);
    }
    return dkey.slice(0, keySize);
}



