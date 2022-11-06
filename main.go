package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type TrackInstance struct {
	Id     string        `json:"id"`
	Name   string        `json:"name"`
	Uptime time.Duration `json:"uptime"`
}
type Result struct {
	Region    string
	Instances []TrackInstance
}

func main() {
	fmt.Println("parse input")
	inputThreshhold := os.Getenv("SEARCH_TIME")
	startRegion := os.Getenv("AWS_REGION")
	//filter := os.Getenv("FILTER")
	if inputThreshhold == "" {
		inputThreshhold = "24h"
	}
	timeThreshhold, _ := time.ParseDuration(inputThreshhold)
	if startRegion == "" {
		startRegion = "us-east-1"
	}
	ctx := context.TODO()
	now := time.Now()
	conf, _ := config.LoadDefaultConfig(ctx, config.WithRegion(startRegion))
	svc := ec2.NewFromConfig(conf)
	fmt.Println("prep")
	regionResponse, _ := svc.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	fmt.Println("collect results")
	results := make([]Result, len(regionResponse.Regions))
	var results2 = make(map[string][]TrackInstance)
	for idx, region := range regionResponse.Regions {
		//fmt.Println(idx)
		results[idx].Region = *region.RegionName
		fmt.Print(".")
		tempconfig, _ := config.LoadDefaultConfig(ctx, config.WithRegion(*region.RegionName))
		descInstanceResponse, _ := ec2.NewFromConfig(tempconfig).DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
		//svc.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
		//fmt.Println(len(descInstanceResponse.Reservations))

		for _, reservation := range descInstanceResponse.Reservations {
			//fmt.Println(len(reservation.Instances))
			for _, instance := range reservation.Instances {
				resInstance := TrackInstance{
					Id:     *instance.InstanceId,
					Name:   getTagValue(*&instance.Tags, "Name"),
					Uptime: now.Sub(*instance.LaunchTime).Round(time.Second),
				}
				//fmt.Println(instance.State.Name)
				results2[*region.RegionName] = append(results2[*region.RegionName], resInstance)
				results[idx].Instances = append(results[idx].Instances, resInstance)
			}
		}
	}
	fmt.Println("\nprint results")
	for _, res := range results {
		if len(res.Instances) > 0 {
			fmt.Println(res.Region)
			for _, ins := range res.Instances {
				fmt.Println(ins.Id)
				fmt.Println(ins.Name)
				fmt.Println(ins.Uptime)
				fmt.Println(ins.Uptime >= timeThreshhold)
			}
		}
	}
	//fmt.Println(json.Marshal(results))
}

func getTagValue(tags []types.Tag, key string) string {
	for _, tag := range tags {
		if *tag.Key == key {
			return *tag.Value
		}
	}
	return ""
}
