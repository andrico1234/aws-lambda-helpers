# EC2 Shutdown Lambda (GoLang)

This is a lambda that is designed to periodically check to see which EC2 instances are running, and to shut them down. 

The Lambda is broken up into these functions 

### Describe Instances

Describe Instance retrieves the data of the EC2 instances. When defining the input we specify that we only want to retrieve the instances whose `instance-state-name` matches what's defined in the `instanceState` constant. Right now it's hardcoded to "running".

Once we get the results back, we retrieve the instances and return them. 

### Format Instances

Format Instances takes in the slice outputted by Describe Instances and normalises it in a way that makes it more simple to traverse, discarding all irrelevant information.

The shape of the data matches the `instanceDetails` struct, which contains the `id`, `keyName`, and `state` of a given instance. We return the formatted list of instances. 

### Update Table

Update Table uses the data outputted by Format Instances to write data to the table. The data will be used as a reference for the corresponding `ec2_startup` lambda. This lambda will check to see the instances that were shutdown by `ec2_shutdown` and start them up again. This ensures that the 2 lambda functions correspond with one another. 

This also means that I'll need to purge the table every time this lambda runs. 

### Shutdown Instances

Shutdown Instances compiles the instanceIds into a single slice, which is then used as the input to the `ec2.StopInstances` function. This then shuts down all of the instances whose Ids were passed. 

## Comments

Feel free to leave comments and suggestions if you feel there are ways to improve the lambda. This is my first major foray into Go programming, so any advice would be appreciated. 
