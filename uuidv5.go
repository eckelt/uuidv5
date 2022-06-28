package main

import "context"
import "log"
import "strings"
import "os"
import "github.com/google/uuid"
import "github.com/aws/aws-sdk-go-v2/config"
import "github.com/aws/aws-sdk-go-v2/service/cloudformation"

func usage() {
	log.Println("USAGE: " + os.Args[0] + " <type> <data>")
	log.Println("	<type> is the prefix of the namespace (e.g. mandant, objekt)")
	log.Println("	<data> is mId + id, or just mId in case of mandant")
}

func readArgs(args []string) (string, string) {
	if len(os.Args[1:]) < 2 {
		usage()
		os.Exit(1)
	}
	return os.Args[1], os.Args[2]
}

func getNamespace(typ string) uuid.UUID {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := cloudformation.NewFromConfig(cfg)

	// Get the first page of results for ListObjectsV2 for a bucket
	output, err := client.ListExports(context.TODO(), &cloudformation.ListExportsInput{})
	if err != nil {
		log.Fatal(err)
	}

	var ns string
	for _, object := range output.Exports {
		if strings.HasSuffix(*object.Name, "-namespace") && strings.HasPrefix(*object.Name, typ) {
			ns = *object.Value
		}
	}
	if ns == "" {
		log.Println("No namespace found for type: " + typ)
		os.Exit(1)
	}

	space, err := uuid.Parse(ns)
	if err != nil {
		log.Println(err)
	}
	return space
}

func main() {
	typ, data := readArgs(os.Args[1:])
	space := getNamespace(typ)
	log.Println(uuid.NewSHA1(space, []byte(data)))
}
