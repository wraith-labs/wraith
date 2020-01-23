from Crypto.Cipher import AES
from Crypto.Protocol.KDF import PBKDF2, HKDF
from Crypto.Hash import HMAC, SHA256, SHA512
from Crypto.Random import get_random_bytes
from Crypto.Util.Padding import pad, unpad

from base64 import b64encode, b64decode
from binascii import Error as Base64Error
from os.path import getsize as file_size
from re import sub as re_sub


class AesEncryption(object):
    '''
    Encrypts data and files using AES CBC/CFB - 128/192/256 bits.

    The encryption and authentication keys 
    are derived from the supplied key/password using HKDF/PBKDF2.
    The key can be set either with `set_master_key` or with `random_key_gen`.
    Encrypted data format: salt[16] + iv[16] + ciphertext[n] + mac[32].
    Ciphertext authenticity is verified with HMAC SHA256.

    Requires pycryptodome https://pycryptodome.readthedocs.io

    :ivar key_iterations: int The number of PBKDF2 iterations.
    :ivar base64: bool Accepts ans returns base64 encoded data.

    :param mode: str Optional, the AES mode (cbc or cfb).
    :param size: int Optional, the key size (128, 192 or 256).
    :raises ValueError: if the mode or size is invalid.
    '''
    def __init__(self, mode = 'CBC', size = 128):
        self._modes = {'CBC': AES.MODE_CBC, 'CFB': AES.MODE_CFB}
        self._sizes = (128, 192, 256)
        self._salt_len = 16
        self._iv_len = 16
        self._mac_len = 32
        self._mac_key_len = 32

        if mode.upper() not in self._modes: 
            raise ValueError(mode + ' is not supported!')
        if size not in self._sizes: 
            raise ValueError('Invalid key size!')
        self._mode = mode.upper()
        self._key_len = int(size / 8)
        self._master_key = None

        self.key_iterations = 20000
        self.base64 = True
    
    def encrypt(self, data, password = None):
        '''
        Encrypts data using the supplied password or a master key.
        
        The password is not required if a master key has been set - 
        either with `random_key_gen` or with `set_master_key`. 
        If a password is supplied, it will be used to create a key with PBKDF2.

        :param data: str or bytes or bytearray The plaintext.
        :param password: str or bytes or bytearray Optional, the password.
        :return: bytes Encrypted data (salt + iv + ciphertext + mac).
        '''
        try:
            data = self._to_bytes(data)            
            if self._mode == 'CBC':
                data = pad(data, AES.block_size)
            
            salt = self._random_bytes(self._salt_len)
            iv = self._random_bytes(self._iv_len)

            aes_key, mac_key = self._keys(salt, password)
            cipher = self._cipher(aes_key, iv)
            ciphertext = cipher.encrypt(data)
            mac = self._sign(iv + ciphertext, mac_key) 

            encrypted = salt + iv + ciphertext + mac
            if self.base64: 
                encrypted = b64encode(encrypted)
            return encrypted
        except (TypeError, ValueError) as e:
            self._error_handler(e)
    
    def decrypt(self, data, password = None):
        '''
        Decrypts data using the supplied password or a master key.
        
        The password is not required if a master key has been set - 
        either with `random_key_gen` or with `set_master_key`. 
        If a password is supplied, it will be used to create a key with PBKDF2.

        :param data: str or bytes or bytearray The ciphertext.
        :param password: str or bytes or bytearray Optional, the password.
        :return: bytes Plaintext.
        '''
        try:
            data = self._to_bytes(data)
            data = b64decode(data) if self.base64 else data

            salt = data[:self._salt_len]
            iv = data[self._salt_len: self._salt_len + self._iv_len]
            ciphertext = data[self._salt_len + self._iv_len: -self._mac_len]
            mac = data[-self._mac_len:]
            
            aes_key, mac_key = self._keys(salt, password)
            self._verify(iv + ciphertext, mac, mac_key)

            cipher = self._cipher(aes_key, iv)
            plaintext = cipher.decrypt(ciphertext)
            if self._mode == 'CBC':
                plaintext = unpad(plaintext, AES.block_size)
            return plaintext
        except (TypeError, ValueError, Base64Error) as e: 
            self._error_handler(e)

    def set_master_key(self, key, raw = False):
        '''
        Sets a new master key.
        This key will be used to create the encryption and authentication keys.

        :param key: str or bytes or bytearray The new master key.
        :param encoded: bool Optional, expexts raw bytes; not base64-encoded.
        '''
        try:
            if not raw:
                key = b64decode(key)
            self._master_key = self._to_bytes(key)
        except (TypeError, Base64Error) as e:
            self._error_handler(e)
    
    def get_master_key(self, raw = False):
        '''
        Returns the master key (or `None` if the key is not set).

        :param raw: bool Optional, returns raw bytes; not base64-encoded.
        :return: bytes The master key.
        '''
        if self._master_key is None:
            self._error_handler(ValueError('The key is not set!'))
        elif not raw:
            return b64encode(self._master_key)
        else:
            return self._master_key
    
    def random_key_gen(self, key_len = 32, raw = False):
        '''
        Generates a new random key.
        This key will be used to create the encryption and authentication keys.

        :param raw: bool Optional, returns raw bytes; not base64-encoded.
        :return: bytes The new master key.
        '''
        self._master_key = self._random_bytes(key_len)
        if not raw:
            return b64encode(self._master_key)
        return self._master_key
    
    def _keys(self, salt, password):
        '''
        Derives encryption and authentication keys from a key or password.
        If the password is not null, it will be used to create the keys.

        :raises ValueError: if neither the key or password is set.
        '''
        if password is not None:
            dkey = PBKDF2(
                password, salt, self._key_len + self._mac_key_len, 
                self.key_iterations, hmac_hash_module = SHA512
            )
        elif self._master_key is not None:
            dkey = HKDF(
                self._master_key, self._key_len + self._mac_key_len, 
                salt, SHA256
            )
        else:
            raise ValueError('No password or key specified!')
        return (dkey[:self._key_len], dkey[self._key_len:])
    
    def _random_bytes(self, size):
        '''
        Creates random bytes; used for IV, salt and key generation.
        '''
        return get_random_bytes(size)
    
    def _cipher(self, key, iv):
        '''
        Creates an AES object; used for encryption / decryption.
        '''
        return AES.new(key, self._modes[self._mode], IV = iv)
    
    def _sign(self, ciphertext, key): 
        '''
        Computes the MAC of ciphertext; used for authentication.
        '''
        hmac = HMAC.new(key, ciphertext, digestmod = SHA256)
        return hmac.digest()
    
    def _sign_file(self, path, key):
        '''
        Computes the MAC of ciphertext; used for authentication.
        '''
        hmac = HMAC.new(key, digestmod = SHA256)
        for data, _ in self._file_chunks(path, self._salt_len):
            hmac.update(data)
        return hmac.digest()
    
    def _verify(self, data, mac, key):
        '''
        Verifies the authenticity of ciphertext.
        
        :raises ValueError: if the MAC is invalid.
        '''    
        hmac = HMAC.new(key, data, digestmod = SHA256)
        hmac.verify(mac)
    
    def _verify_file(self, path, mac, key):
        '''
        Verifies the authenticity of ciphertext.

        :raises ValueError: if the MAC is invalid.
        '''
        hmac = HMAC.new(key, digestmod = SHA256)
        beg, end = self._salt_len, self._mac_len

        for chunk, _ in self._file_chunks(path, beg, end):
            hmac.update(chunk)
        hmac.verify(mac)
    
    def _error_handler(self, exception):
        '''
        Handles exceptions (prints the exception by default).
        '''
        raise exception
        # Do not print. We want to be able to catch the exception
        #print(exception)

    def _file_chunks(self, path, beg = 0, end = 0):
        '''
        A generator that reads a file and yields chunks of data. 
        The chunk size should be a multiple of 16 in CBC mode.
        '''
        size = 1024
        end = file_size(path) - end

        with open(path, 'rb') as f:
            pos = (len(f.read(beg)))
            while pos < end: 
                size = size if end - pos > size else end - pos
                data = f.read(size)
                pos += len(data)

                yield (data, pos == end)
    
    def _to_bytes(self, data, encoding = 'utf-8'):
        '''
        Converts unicode strings and byte arrays to byte strings.
        '''
        if hasattr(data, 'encode'):
            data = bytes(data, encoding)
        if type(data) is bytearray:
            data = bytes(data)
        return data


