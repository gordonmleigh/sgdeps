package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/rds"
)

func main() {
	var region, securityGroup string

	flag.StringVar(&region, "region", "", "the AWS region")
	flag.StringVar(&securityGroup, "sg", "", "the id of the security group")
	flag.Parse()

	if securityGroup == "" {
		fmt.Println("Expected security group id")
		os.Exit(2)
	}

	config := &aws.Config{}
	if region != "" {
		config.Region = &region
	}

	sess := session.Must(session.NewSession(config))

	ec2svc := ec2.New(sess)
	inboundGroupId := "ip-permission.group-id"
	fmt.Println("Security Groups (inbound rules):")

	groups, err := ec2svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   &inboundGroupId,
				Values: []*string{&securityGroup},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, group := range groups.SecurityGroups {
		fmt.Printf("  - %s\n", *group.GroupId)
	}

	outboundGroupId := "egress.ip-permission.group-id"
	fmt.Println("\nSecurity Groups (outbound rules):")

	groups, err = ec2svc.DescribeSecurityGroups(&ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   &outboundGroupId,
				Values: []*string{&securityGroup},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, group := range groups.SecurityGroups {
		fmt.Printf("  - %s\n", *group.GroupId)
	}

	groupId := "group-id"
	fmt.Println("\nNetwork Interfaces:")

	enis, err := ec2svc.DescribeNetworkInterfaces(&ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			{
				Name:   &groupId,
				Values: []*string{&securityGroup},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, eni := range enis.NetworkInterfaces {
		fmt.Printf("  - %s\n", *eni.NetworkInterfaceId)
	}

	instanceGroupId := "instance.group-id"
	fmt.Println("\nEC2 Instances:")

	instances, err := ec2svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   &instanceGroupId,
				Values: []*string{&securityGroup},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			fmt.Printf("  - %s\n", *instance.InstanceId)
		}
	}

	fmt.Println("\nLoad Balancers:")
	elbSvc := elb.New(sess)

	loadBalancers, err := elbSvc.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	if err != nil {
		log.Fatal(err)
	}

	for _, loadBalancer := range loadBalancers.LoadBalancerDescriptions {
		for _, group := range loadBalancer.SecurityGroups {
			if *group == securityGroup {
				fmt.Printf("  - %s\n", *loadBalancer.LoadBalancerName)
				break
			}
		}
	}

	fmt.Println("\nRDS:")
	rdsSvc := rds.New(sess)

	dbs, err := rdsSvc.DescribeDBSecurityGroups(&rds.DescribeDBSecurityGroupsInput{})
	if err != nil {
		log.Fatal(err)
	}

	for _, dbsg := range dbs.DBSecurityGroups {
		for _, group := range dbsg.EC2SecurityGroups {
			if *group.EC2SecurityGroupId == securityGroup {
				fmt.Printf("  - %s\n", *dbsg.DBSecurityGroupName)
				break
			}
		}
	}
}
