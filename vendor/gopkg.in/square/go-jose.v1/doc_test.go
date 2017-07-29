/*-
 * Copyright 2014 Square Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package jose

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

// Dummy encrypter for use in examples
var encrypter, _ = NewEncrypter(DIRECT, A128GCM, []byte{})

func Example_jWE() {
	// Generate a public/private key pair to use for this example. The library
	// also provides two utility functions (LoadPublicKey and LoadPrivateKey)
	// that can be used to load keys from PEM/DER-encoded data.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Instantiate an encrypter using RSA-OAEP with AES128-GCM. An error would
	// indicate that the selected algorithm(s) are not currently supported.
	publicKey := &privateKey.PublicKey
	encrypter, err := NewEncrypter(RSA_OAEP, A128GCM, publicKey)
	if err != nil {
		panic(err)
	}

	// Encrypt a sample plaintext. Calling the encrypter returns an encrypted
	// JWE object, which can then be serialized for output afterwards. An error
	// would indicate a problem in an underlying cryptographic primitive.
	var plaintext = []byte("Lorem ipsum dolor sit amet")
	object, err := encrypter.Encrypt(plaintext)
	if err != nil {
		panic(err)
	}

	// Serialize the encrypted object using the full serialization format.
	// Alternatively you can also use the compact format here by calling
	// object.CompactSerialize() instead.
	serialized := object.FullSerialize()

	// Parse the serialized, encrypted JWE object. An error would indicate that
	// the given input did not represent a valid message.
	object, err = ParseEncrypted(serialized)
	if err != nil {
		panic(err)
	}

	// Now we can decrypt and get back our original plaintext. An error here
	// would indicate the the message failed to decrypt, e.g. because the auth
	// tag was broken or the message was tampered with.
	decrypted, err := object.Decrypt(privateKey)
	if err != nil {
		panic(err)
	}

	fmt.Printf(string(decrypted))
	// output: Lorem ipsum dolor sit amet
}

func Example_jWS() {
	// Generate a public/private key pair to use for this example. The library
	// also provides two utility functions (LoadPublicKey and LoadPrivateKey)
	// that can be used to load keys from PEM/DER-encoded data.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Instantiate a signer using RSASSA-PSS (SHA512) with the given private key.
	signer, err := NewSigner(PS512, privateKey)
	if err != nil {
		panic(err)
	}

	// Sign a sample payload. Calling the signer returns a protected JWS object,
	// which can then be serialized for output afterwards. An error would
	// indicate a problem in an underlying cryptographic primitive.
	var payload = []byte("Lorem ipsum dolor sit amet")
	object, err := signer.Sign(payload)
	if err != nil {
		panic(err)
	}

	// Serialize the encrypted object using the full serialization format.
	// Alternatively you can also use the compact format here by calling
	// object.CompactSerialize() instead.
	serialized := object.FullSerialize()

	// Parse the serialized, protected JWS object. An error would indicate that
	// the given input did not represent a valid message.
	object, err = ParseSigned(serialized)
	if err != nil {
		panic(err)
	}

	// Now we can verify the signature on the payload. An error here would
	// indicate the the message failed to verify, e.g. because the signature was
	// broken or the message was tampered with.
	output, err := object.Verify(&privateKey.PublicKey)
	if err != nil {
		panic(err)
	}

	fmt.Printf(string(output))
	// output: Lorem ipsum dolor sit amet
}

func ExampleNewEncrypter_publicKey() {
	var publicKey *rsa.PublicKey

	// Instantiate an encrypter using RSA-OAEP with AES128-GCM.
	NewEncrypter(RSA_OAEP, A128GCM, publicKey)

	// Instantiate an encrypter using RSA-PKCS1v1.5 with AES128-CBC+HMAC.
	NewEncrypter(RSA1_5, A128CBC_HS256, publicKey)
}

func ExampleNewEncrypter_symmetric() {
	var sharedKey []byte

	// Instantiate an encrypter using AES128-GCM with AES-GCM key wrap.
	NewEncrypter(A128GCMKW, A128GCM, sharedKey)

	// Instantiate an encrypter using AES256-GCM directly, w/o key wrapping.
	NewEncrypter(DIRECT, A256GCM, sharedKey)
}

func ExampleNewSigner_publicKey() {
	var rsaPrivateKey *rsa.PrivateKey
	var ecdsaPrivateKey *ecdsa.PrivateKey

	// Instantiate a signer using RSA-PKCS#1v1.5 with SHA-256.
	NewSigner(RS256, rsaPrivateKey)

	// Instantiate a signer using ECDSA with SHA-384.
	NewSigner(ES384, ecdsaPrivateKey)
}

func ExampleNewSigner_symmetric() {
	var sharedKey []byte

	// Instantiate an signer using HMAC-SHA256.
	NewSigner(HS256, sharedKey)

	// Instantiate an signer using HMAC-SHA512.
	NewSigner(HS512, sharedKey)
}

func ExampleNewMultiEncrypter() {
	var publicKey *rsa.PublicKey
	var sharedKey []byte

	// Instantiate an encrypter using AES-GCM.
	encrypter, err := NewMultiEncrypter(A128GCM)
	if err != nil {
		panic(err)
	}

	// Add a recipient using a shared key with AES-GCM key wap
	err = encrypter.AddRecipient(A128GCMKW, sharedKey)
	if err != nil {
		panic(err)
	}

	// Add a recipient using an RSA public key with RSA-OAEP
	err = encrypter.AddRecipient(RSA_OAEP, publicKey)
	if err != nil {
		panic(err)
	}
}

func ExampleNewMultiSigner() {
	var privateKey *rsa.PrivateKey
	var sharedKey []byte

	// Instantiate a signer for multiple recipients.
	signer := NewMultiSigner()

	// Add a recipient using a shared key with HMAC-SHA256
	err := signer.AddRecipient(HS256, sharedKey)
	if err != nil {
		panic(err)
	}

	// Add a recipient using an RSA private key with RSASSA-PSS with SHA384
	err = signer.AddRecipient(PS384, privateKey)
	if err != nil {
		panic(err)
	}
}

func ExampleEncrypter_encrypt() {
	// Encrypt a plaintext in order to get an encrypted JWE object.
	var plaintext = []byte("This is a secret message")

	encrypter.Encrypt(plaintext)
}

func ExampleEncrypter_encryptWithAuthData() {
	// Encrypt a plaintext in order to get an encrypted JWE object. Also attach
	// some additional authenticated data (AAD) to the object. Note that objects
	// with attached AAD can only be represented using full serialization.
	var plaintext = []byte("This is a secret message")
	var aad = []byte("This is authenticated, but public data")

	encrypter.EncryptWithAuthData(plaintext, aad)
}
