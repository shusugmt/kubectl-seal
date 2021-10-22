package cmd

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"time"

	"github.com/bitnami-labs/sealed-secrets/pkg/crypto"
	"github.com/spf13/cobra"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

type genkeyCmdOptions struct {
	keySize  int
	validFor time.Duration
	myCN     string
}

var genkeyCmdOpts = &genkeyCmdOptions{}

func init() {
	// https://github.com/bitnami-labs/sealed-secrets/blob/537244be6d4d44d300dd07e47b599e4bd0ba4b71/cmd/controller/main.go#L42-L44
	genkeyCmd.Flags().IntVar(&genkeyCmdOpts.keySize, "key-size", 4096, "size of encryption key")
	genkeyCmd.Flags().DurationVar(&genkeyCmdOpts.validFor, "key-ttl", 10*365*24*time.Hour, "duration that certificate is valid for")
	genkeyCmd.Flags().StringVar(&genkeyCmdOpts.myCN, "my-cn", "", "common name to be used as issuer/subject DN in generated certificate (default \"\")")
}

var genkeyCmd = &cobra.Command{
	Use:   "genkey",
	Short: "generate a new sealing key pair",
	Long:  `Generate a new sealing key pair.`,
	Run: func(cmd *cobra.Command, args []string) {

		key, cert, err := crypto.GeneratePrivateKeyAndCert(genkeyCmdOpts.keySize, genkeyCmdOpts.validFor, genkeyCmdOpts.myCN)
		if err != nil {
			log.Fatalf("%v", err)
		}

		certPEM := pem.EncodeToMemory(&pem.Block{Type: certutil.CertificateBlockType, Bytes: cert.Raw})
		keyPEM := pem.EncodeToMemory(&pem.Block{Type: keyutil.RSAPrivateKeyBlockType, Bytes: x509.MarshalPKCS1PrivateKey(key)})
		fmt.Printf("%s%s", certPEM, keyPEM)
	},
}
