package signer

import (
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/config"
	"github.com/kruftik/jenkins-update-dot-json-resigner/internal/jenkins/types"
)

func TestSigner(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	s, err := NewSignerService(logger.Sugar(), config.SignerConfig{
		CertificatePath: "../../../testdata/certs/test.crt",
		KeyPath:         "../../../testdata/certs/test.key",
	})
	if err != nil {
		t.Fatal(err)
	}

	unsigned := &types.InsecureUpdateJSON{
		Plugins: map[string]types.Plugin{
			"text": {
				URL: "http://origin.local/download/plugin.hpi",
			},
		},
		UpdateCenterVersion: "123",
	}

	validSignature := types.Signature{
		Certificates: []string{
			"MIIFJTCCAw2gAwIBAgIUO6LOosiA3ZxN+MumSDSWu6GcI0QwDQYJKoZIhvcNAQELBQAwIjEgMB4GA1UEAwwXWW91ckplbmtpbnNVcGRhdGVDZW50ZXIwHhcNMjQwODE4MDk0NzA5WhcNMzQwODE2MDk0NzA5WjAiMSAwHgYDVQQDDBdZb3VySmVua2luc1VwZGF0ZUNlbnRlcjCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBANHghc3KYzgKOSIOt5ahc6VY8mHRannga6o8inlF3xchDLcuRnvJSJwTwSTsgzXDvfICr1KI2+0RTkFCwhK66Px4KTd2+/zWT+xgrTGGGixGx0J/m+y8YuceczEOs60LAcuh47BXBgihEL638H3b6u6pWwZQD0awilC2FboB7stt5WYmqv3qbnJGTbFhSWkgPaloVFRNd1cx9CuvfvaZqA3+gMCRB4joAd09CAavQFL+l15jZkgogoslnk2bNFxCzEawgnLk4pJXZ04vmfdry469B/iWMveZbAgg3hmUjPR1mfOmpNjGz8TqJT37HkiVqBU8ixwWDZkOpsaDTv6SBEISP9ZiGFxTQ8MQpwbpVmQVJDEwoXR0j4p5FDVYyZD9MKsjgXrWPbmx1LLMHdh8N6DXLUEAlyHgfKSh9yHxvfMP7XyJGd/YHfa/0XN6CFRkF8CrQ6gbO3Cr1A2cK+v1P/NNubVDlu30OcXe+e/be7+S3RIjcyCG61VEFql1fbgULymqNlCm9lZxXr5WA3g7Vdr7BGvhHL3VXydD1O4lf6BvQhCZ8+06YelY6Wa/xMb2nb+hCGY9ofCaPFpzC/Vo/AIi0pFJyTg7odS+CH2JDpxhn0WIFJk6b+GhunElR5zsUu4Pu7nbnIxqU464fLxLMx8HTf80WGU0ntcaWZV7oBlXAgMBAAGjUzBRMB0GA1UdDgQWBBTWwk1sEkoO230VbNQdQq34GQfflDAfBgNVHSMEGDAWgBTWwk1sEkoO230VbNQdQq34GQfflDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4ICAQA9MCivKR3a8EEDgaElA60y+S8U/buGQyG2sGGQnvNP6PSXrgB75IdU6HK+LBikgtKkypzHniUsNIrgsdfshMlrJk0KuZbleRSTbXa4I6eteRDnpF5wvQNUGUIpImNXt9vSsmjPuZZacGgQWR0/DkPHlB2qf0EtzEXYGpWzR6UJaStBHu3q5A+OxAJSefU14MtQmD/egx4HB1faWQ1qz2EBUBwg5iqnvBdPmBuNAbrz1tDWZxDbacKdSPi0GmFuur0n0UpsDL94HDuEELKXGEkuPg7XJ/VHEPooaIVd5LvoKfVu0kuUV9oZ5qmxx3nMp1EqnAVy97lLkgF+9EL6wBmiBR6rL0UKYSyd+WSp4JeA1SHj9WOr6WRoOLaY0ajNNqjEKuXsugoaDZtD9hWS0Zb0ioXQcFjatWgOrm9FK9tYu3GGdp77k3yVjcblyLugkXrnnXQgKHRBQGELODM1SpB6tKEGlS2uG8AzqC5+FefoobImBzs3Mhnebvpl5urwzHe1QEMtudJq9VqoobKgecLXESnKmDJ1LasYz1zs+9cUPL07INIE3mSdlZ9qvIKBSnC2ycsKFV5nWr4QEPzpVJUQ01FGKjfQPEe3uuJEX0ntF3VmOQATcs36gY8+U7X7qwa6HxssFskhLRVYMvMxzYhpgIaHAekU2mbvw/kvSakXXQ==",
		},
		CorrectDigest:       "59fd7b4e6de2c7b4859f293b0c0e9c540376c96a",
		CorrectDigest512:    "btNK63Unffc+nLEOEbB9/3OXX2+jr1KdYYfH9Rd+TmYTyDkM05r4JBailigV6KwHmEFLDxiVNzRxIkPY7zGnWw==",
		CorrectSignature:    "EsooHadltZ8hyppA5Ofats5ToQqBauXXeBUB6m5v8dhtWJw0skk4UrGfxQJWsQ3n9JyuYZHDe5YQq7PdEBX9m7exAfabRvS+Hh4zOBASBv78/D+mRVCr+UM6rbSU2bniOubj61WoPurJQbJhZP9RUAXiEdX556dOCanm9Ptfopc6gyODsx2R1CqkzyyVz9jvDjKDv6U87osybmBxFk09VQhm7K7TX4T0aYy20QIF3R7tcHNFAwYGiULXHTP9GLFLw8+s43LlyF0nzatdGOrPgUugRhEn+Pud6AErlgInUcQycR6JLoD/F9mwKNC8tNg77Yn282+2/ULP09iozfbTJAXxExCUE3YRnC2piUr/gDJN7NcyIl0tr/9qmUqJBsIl80f9av6jPN5vKQG12cYPV2Z1sJsMFstAUaUfi+pdMXNOab8aVka6foxflM4ywBUUFH2cm/l7ktjCVXR6SO/3KnI6iS75ZGLE0sNMBslOYKGqXZiGkVOzfQkE7AQIdPNIj+gGXNcWWmjhgLA9xsAx6vtpNNwrsCRLStkviErmZL7qKc/ODvU8WOcx7qVzeQVnN6fbOxrs/eDKZgN0/pZmrkkB4HCo6/SZu4vzP5KvVHj4anlBt8OMT99Ziy7ArhzPk1mvlX8eoFanLKwFXPTRJVygDw9Cu66n049xQKK20sk=",
		CorrectSignature512: "2674d0326c8e12d5e1844ebaf73ea82a6ccd7c339f69e2b01bcc72d771db8753d2e36ed76510544d33522e2f0c6973b89851dae85552e296c519c2b047fd50bab8cdefc7373c69a696e895546295bb4fd216aa49d75eaf65e80ac1001f3c498cd78cd14c0516c8196acc1effd36575fd7ca13ead3e83e9a2da44ff70df4069c8189d4e1a38ef8cc7f668ce6a4f886d7b78ccfeaddfdd106605ef4b154c7fb55434b1dd8812e06d770dfb6047a0a4f5848db5cd2ed85902d03a970f83be09ed8ffda8eca0bac619439a16238006027e3a5f1a87d92c22d1e5ca9446392246b53d2c6499ebd3a638247f197dac4ac9f30901e92914ed5d5b7084b59a3626113144b367fba1cc9b18490f1ad1c4bcb07b9ed02b8b185f7873bcb23e8b5cf416632998ef1c3aca8a2a59a0ba2538c22e83f55b0ca56606cf859c16853a7de1b55758fab72d9555fc0bf5ec691b422959405385a57a4e7fcaa49202d2330c5180d09c3ecc2a70e60ceb63cf63c71721f62a82aceb0cf5286a7232ec63ddb0404ca7796974a8c2d6b909f67feda74b1a8eee733b56ec8e697b2c32467890d63e0b9499908b87bc1daeb2a1bd7b46b3ca418dabe58cb58b30f6f78e4cb2ada03bd3c04066dc8c4c4e9d004d94f12d3fd1c6c9a99ad351c98163fbb7a8f35cb41713d25f4d1dfdb49423572ca9656386ab2e451a36705180b501fb17bdf50878146430af",
	}

	if err := s.VerifySignature(unsigned, validSignature); err != nil {
		t.Fatal(err)
	}

	signed := types.SignedUpdateJSON{
		InsecureUpdateJSON: unsigned,
	}

	if err := signed.Sign(s); err != nil {
		t.Fatal(err)
	}

	newValidSigrature := types.Signature{
		Certificates: []string{
			"MIIFJzCCAw+gAwIBAgIUPgn3DpcF26y1iRdBQqXRH4U9bwcwDQYJKoZIhvcNAQELBQAwIzEhMB8GA1UEAwwYWW91ckplbmtpbnNVcGRhdGVDZW50ZXIyMB4XDTI0MDgxODEwMzc0MloXDTM0MDgxNjEwMzc0MlowIzEhMB8GA1UEAwwYWW91ckplbmtpbnNVcGRhdGVDZW50ZXIyMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAwUUxvwmc7l0ELD6HHwKfJHxqNOcUbTBF5xKwwKIWjK6g69TcBZ9Jr/jbQGyspw5dfJoJ/iGlSUmM6O2F7EkSI4h/Wx+TzuhxU+bHcqRMIgdYrxz+48u/a0UoYexmTUHAOnwE8Y4U+xMHi7QaIqUUarQoAct8Al+7salkXqFTcOWwgErgZ271zlxs0P1w7EWbG5hkFOC0pG4tW2bezk2/NOLZQb+ir24CTspwDUUn79olq7dnBiiVuPu6qwHXq9zb3U6BwSLue0g3LdVRHUBLwQDsaI5kiia2grbF5DKE48/eE4zLyM8f6QZha4HaZ/uJlsqrNXGJG3cXZOyT5b0RuDHFe3BfvHHE7a2YtcWyj1fysDHZ7NX50vk8pr6LybyKcWpJPizrA2yCusiLzMBmHb/JERAAuB1a6ajRMFY4w/cTSdCoRtQCiL3Xf/0SKdefDLNmPtk7a+ZDs6ZYvS81iYZKHVjoj9JuJCzVRWW1HtI+1++TTcCBcUJG/qNZXrtoT22lPGMzv9Jw9Z9RNIBXtCFvo0AdpGGF44vnl0EmFhPiPM2+mT/ptsg36Cc35EtvmuDg5G+BFaiH/mG7u+gsHCuQNwi2DXtg3XP9UszSCR/rPTF4QDduDSK4sXIoPKRsdmHglfbW4A28LFCPV3Gqulcavr3PJ5v0wZ8FTsvHkuMCAwEAAaNTMFEwHQYDVR0OBBYEFDKia34aS2ZsTyLK8vkPUjcuoCAjMB8GA1UdIwQYMBaAFDKia34aS2ZsTyLK8vkPUjcuoCAjMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggIBAGSsNYcaNT+bang++Rllws/UrdZOJqo31REii2ePEUSs9z2SxNZHkxhK0x+i2pQKqK3Z49qeD5qorEZq+56fheU3g5slkMlFLm3nfn7Wi+o6GmkH9biHOH+h/kDNVMy6n2K/y1g7US8tzjoRImZmJSqyh6ggmEQPp5Q2DU1G9TJX9bNCvGUVX11vFPpBokwIEi6CigTT7wK3fPMeDXTNMYjX+lqkIDVd30WI3lzrHhbNFFfYbk+OpqdrytKojX0sFcl0LwUcU//nqc4Q7/3JEHwYHPhnatWY+/zRgnq1Qk4lnN++lu3OCVZAPHfOKfGkQkBcVx+91e4+/My5AnO1dGgEQ6LJkWWHBi7M1F9ybFEfuHNz+oMlq3hkUeF4foSD6lSvV+9IIU5vRGBJKmodwrfQUTD+EoI78Qq7qU44EtXJK39GahHPzas1WdvYx/5ZjoQYDncgeOVvj2c+GK54MZHAPkrHbfFVxgHvyXPJKXqH8RniOnAVtewjKawyva/uXvwyG4STziSz5PfSS8DPhtgvH4Rq/f/btqBWYBt0F/XkJt2AxcjTklryWYWjiZ+F9uSOKH45w0suaSKA/ovBv6Sf9F7qTBv1yqgvnUROrbcjtPd1bnrN+FwLeVgePxYFkovvFdLhPgwVJSHM/aJnnJeHFZ6I9sRLwOIhKdGTUbsO",
		},
		CorrectDigest:       "59fd7b4e6de2c7b4859f293b0c0e9c540376c96a",
		CorrectDigest512:    "btNK63Unffc+nLEOEbB9/3OXX2+jr1KdYYfH9Rd+TmYTyDkM05r4JBailigV6KwHmEFLDxiVNzRxIkPY7zGnWw==",
		CorrectSignature:    "FBqLNqw0m4mn6WqcPbZ9FvqkPet22DXm/oZwkbRdZHJ8QYIo3D93PaTOVlMnYQT4Xcq5Bk7zRhyqY5ocP0ivOFK+1x+Mi3Mi9aviZegfDZZR01V3KIiQRPkR+lMB2Fqm98zJ7UnyQq1wZmTkdalek8ndgSjKQJ74pmxvUhxGDJ37D2Mjm5Tc+7lJTNU+uIhq4XNUidsakPBEG9qnCqd3Cd78KRtzVwkLB0dytdPIxzQvFN0wEq1rUSZInQ/gZy+eGa/V9EmnEk1NGI3LSdtPQ1SJ4j6HGbvDKTLFnu2GkaPrSlOIStldH5FDF/fA3C0hi9lz5+1X08D/wA/mYR80pIna0BeKOvjTexmL4kqcaAe+yk1+5ckN4/0QPaAz0w24B4OYtmA1p5xL+wYAV6P7VehvZnC/kv4J6ShP1fxJtsQUX/ygIn6EzEMv7Hk6ONMiemHA3J//qgWNMzVB3tK8oHOI0UulhfYmO+s+IHMQvJzvo5FRgY/E56GDzdvxrmlVzt5ZR5W5y+hGLN2pVFpc8poszyw7WVGX35aC42hYwVg4w3ZXBPPcqj+9TznWlih6g+5derqHl3etLrZB/eAFDW7FEHD0KLL/AI5+Kzgf3+NNtxWFvRpn4PwPs941bgG/H/ucLrcxfdhtmj0Um4xywTJan/jMN3D+gAe13ozLOcA=",
		CorrectSignature512: "011ce4b1fe757dc5e544feaa59a8f7b31f1a3eca2e1bb83d1c08ca8f6481460242faf223387b327585b348cb1bbf4815dba0d15594f4fea0ae232cb0fa353ce9c9da5522cc688a24eec5ed84840e43cc5796a16452c2f076c9690370ead4c31c5954e16eb83ddfccd5c3b715faff2f3958433915bc0a480fcc1d1348a194b5ba03d5b49833a51e515ff40ea5686c787d37f417aac47bb5174aaa788983ba963c7d068660796d9a07ddd5e8a8716772b79e3cec492a078c6d3780da1375b535c0142f478bae985352b76cd6c8b7e3018ae8c4c99dfcc62866fe1d3fd87d644f8e73a6ce092d840bcb63831c0b3109f15633fd7b7f340b42e3666d15a270c9a6d95b3e5eae742da2b933acce0e7176593011f5ce0726192f4f990ede5ba312367825f022419d66bf3687c467a529d70717ccc258caa9da91f0a01315b26e418ebce1077760a13047bf2a67359b069ffd4887231299bedb559ab999884837c91cb8d6ccc40b05b0c1f2bad02036ecdbbe496c6daaa254b5abaf76c80f4d0fcd5dca953a6cd49a5e190721455078733e69246e13775c8459ad71c839415f980e4b1e4d3dbe7af68a2fd595c261b0c636981457ae038fd26cb5eabb5fa70fbcf8f6a05264f427035c1b4486ad5adf758eecc4190be62a531d0652798aa1c5a88d59d842836a91e373891fcb242a8f22211d35017f9c7e4989dcf48892bdf45b3230c3",
	}

	if !reflect.DeepEqual(signed.Signature, newValidSigrature) {
		t.Fatal("Signatures do not match")
	}
}