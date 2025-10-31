I need you to design and develop a CLI tool in golang (+ any other language that you think would be more useful for specific use cases).

## Purpose of the Ark CLI tool

To automate the most useful and complex things for developers, to provide an easy to use interface that provides:

- Best User Experience
- Industry grade automation
- Industry grade security
- Extensibility for the future features and automations
- Customizations and configurations so that developers can configure the tool as per their need.

## Allowed Tech Stack

- Golang as the main language for the CLI, also the main entry point: Because we need independent cross platform compatibility (for linux + macos)
- Databases: that can run natively on the OS without requiring the full blown setup, but should have some security compatibilities.
- Additional Tech Stack Allowed: Any other language that can be compiled with the main binary and can run native operations better than golang (if the specific features are not supported by golang)

## Features and Functional Requirements

### On Device Automations

- caffeinate feature, to keep the device up and running (preventing from sleep) by moving the cursor with specific time interval (not dependent on the OS APIs, because those are un-reliable)
- credentails and other sensitive data store: Like vault, storing any kind of configuration safely and in encrypted format with industry grade security: So that the user can store the credentials directly and can directly copy or read them whenever needed. (features like, list, search, get, update, delete should be supported here. + formatting is a must so that user can properly retrieve the secrets in specific formats (as they have stored them like JSON, YAML, String or something like that))
- directory locking: The tool shold allow user to lock specific directories (with master password), so that no one else can access the content within the directory (optionally hide locked directory could also be supported)

### Integration With Cloud

The tool should support the cloud services based features (for now lets just focus on the AWS)

First thing we need the ARK to do is perform pre-req checks, like if the AWS CLI is installed or not, the credentials are configured or not.
Then It should allow specific operations using the AWS CLI (ark working as a wrapper around the AWS CLI), allowing easily perform operations such as:

- list AWS accounts
- Select the AWS account or specify default AWS account (so that ark usages that profile)
- Test the connection and credentials (to check if the ark has access to that account, by running some test command with aws cli and those credentails)
- List EC2 instances
- List S3 Buckets

#### EC2 Instances

Ark should allow registering the EC2 instances with custom names and the ssh key.

- So that develoeprs can quickly shutdown, or power on those instances directly from ark (using AWS CLI)
- Get the EC2 performance metrics (CPU, RAM, Storage statics, and usages)
- Quick SSH to EC2 instance

Any other userful thing that you can think of here

#### S3 Buckets

- List Buckets
- Navigate in bucket
- Quick Upload and Download

And for all these the ark should store the sensitive information within itself (with the vault feature) so the developers don't have to worry about providing credentials, ssh keys etc again and again.

## Backup & Restore

The ark should be allowed to create and store its own backup (in encrypted form) within S3 bucket. It should provide the options to configure the S3 bucket (either choosing an existing one or creating a new one) and the Ark should have functionality to restore point it time from backup (when provided the configuration like s3 bucket name etc, then the ark can list down stored backups on that s3 bucket and allows user to choose a backup to restore from)

For now we need only these features but going forward we might implement more and more features within the ark.

## Non functional requirement:

### Project Structure

The directory structure should be intutive so that going forward if a developer need to understand and debug the issue, it is easy to navigate through features and domains of features. We need properly nested hierarchical structure. Also the code should be modular.
The code should be optimized and robust and defensive (with proper error handling and logging). It should allow us to plugin more features going forward in the future with easy without messing up the whole structure or requiring a lot of efforts.

### User Experience

User experience is a top priority for us, we don't want the developers to be pulling their hairs out while using this tool, the tool should be so easy to use that the developers tend to use it more and more.
The command line interface should be interactive, with proper formatting and color coding and in place updation (like whatever are the latest features there in the world for the CLIs we need those) we need a world class interface here.

### Installation

The tool should be easy to install with make file.

- It should install the tool, put it in the bin and update the path in specific file (based on the terminal or shell being used, like bashrc or zshrc file)

### Uninstallation

The tool should provide complete uninstallation with single command, without leaving any trace. And when the tool is uninstalled it should unlock the directories and unhide the hidden directories.

### Logging

- The tool should maintain the logs (activity logs + the application logs) with rotation mechanism so it doesn't keep filling the disk space forever (logs should be allowed to rotate every day, and purging the old log files should be automated)
- The tool should have its own view logs command (with specific features like view logs for this feature or that feature) so that we don't need to go and find those specific log files and read them. The tool should provide those features by itself.

### Priorities:

- Currently we need the tool to work 100% on the macos (silicon chip), while keeping the space for the other OS implementation later (wherever the other OS specific functionalities are not implemented it should just print out the functionality is coming soon)
- We should implement all the above listed features right now. Feature wise there is nothing that we don't want right now.

Git Repository for this project: https://github.com/mbeniwal-imwe/ark.git
