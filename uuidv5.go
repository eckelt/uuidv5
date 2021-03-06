package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/eckelt/uuidv5/awsExports"
	"github.com/eckelt/uuidv5/s3Data"
	"github.com/google/uuid"
)

func usage() {
	fmt.Println("USAGE: " + os.Args[0] + " [-n namespace] [-c data] [-b mId-uuid uuid]")
	fmt.Println("	-n namespace 	namespace to convert the uuids into or from (e.g. personen, objekte)")
	fmt.Println("	-c <data>	convert given data mId+id into uuid in given namespace")
	fmt.Println("	-b <mId> <uuid>	looking for mandantId and uuid in given namespace")
	fmt.Println("	-h		this usage")
}

func readArgs(args []string) (namespace, data, mId, uuid string, verbose bool) {
	verbose = false
	for i := 0; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-n":
			i++
			namespace = os.Args[i]
		case "-c":
			i++
			data = os.Args[i]
		case "-h":
			usage()
			os.Exit(0)
		case "-v":
			verbose = true
		case "-b":
			i++
			mId = os.Args[i]
			if len(os.Args) > i+1 && len(os.Args[i+1]) == 36 {
				i++
				uuid = os.Args[i]
			}
		}
	}
	return
}

func keys(m map[string]uuid.UUID) []string {
	//The default length of the array is the length of the map. When the array is attached, there is no need to re apply for memory and copy, which is very efficient
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func main() {
	ctx := context.Background()

	ns, data, mId, needle, verbose := readArgs(os.Args[1:])
	if ns == "" {
		fmt.Fprintln(os.Stderr, "Parameter -n namespace is mandatory")
		usage()
		os.Exit(2)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading config:", err)
		os.Exit(1)
	}

	exports := awsExports.NewFromAwsConfig(cfg)

	namespaces, err := exports.Namespaces(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// generate uuid for given type
	nsUuid, ok := namespaces[ns]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown namespace \"%s\". Do you mean any of these: %s\n", ns, strings.Join(keys(namespaces), ", "))
	} else {
		if data != "" {
			fmt.Println(uuid.NewSHA1(nsUuid, []byte(data)))
		}
	}

	if mId != "" {
		// backwards search for given uuid
		stage, err := exports.Stage(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if verbose {
			fmt.Println("Stage:", stage)
		}

		s3Data := s3Data.NewFromAwsConfig(cfg, stage, namespaces)

		mandantId, err := s3Data.Mandant(ctx, mId)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if verbose {
			log.Printf("Checked mandanten in s3 and found %s\n", mandantId)
		}
		if needle == "" || ns == "mandanten" {
			fmt.Println(mandantId)
		} else {
			uuids, err := s3Data.Rainbow(ctx, mandantId, ns)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if verbose {
				log.Printf("Created rainbow table of %d entries of %s for this mandantor\n", len(uuids), ns)
			}
			found, ok := uuids[needle]
			if !ok {
				fmt.Printf("%s not found in %s for mandant %s\n", needle, ns, mandantId)
			} else {
				fmt.Printf("%s %s\n", mandantId, found)
			}
		}
	}
}
