package main

var (
	AttackActionNone           = "none"
	AttackActionOther          = "other"
	AttackActionIdentification = "identification"
	AttackActionHoneypotCheck  = "honeypot_check"
	AttackActionDropBinary     = "drop_binary"
	AttackActionAddSSHKey      = "add_ssh_key"
)

var attackActions = []string{
	AttackActionNone,
	AttackActionOther,
	AttackActionIdentification,
	AttackActionHoneypotCheck,
	AttackActionAddSSHKey,
	AttackActionDropBinary,
}

func classifyAction(attack *Attack) string {
	act := AttackActionNone
	if attack.Classification == "command_entered" || attack.Classification == "ssh_command" {
		act = AttackActionOther
		if containsAny(attack.Contents, []string{
			"chmod 777 .",
			"echo -ne \"\\x7f\\x45\\x4c\"",
			"; chmod u+x ",
			"curl -O http://",
			"curl -s -L -O ",
			"ftpget -v -u anonymous -p anonymous",
		}) {
			act = AttackActionDropBinary
		} else if containsAny(attack.Contents, []string{
			">> authorized_keys",
			">.ssh/authorized_keys",
			"echo \"ssh-rsa AAAA",
			"|chpasswd|bash",
		}) {
			act = AttackActionAddSSHKey
		} else if containsAny(attack.Contents, []string{
			"nc localhost 12",
			"# ls -lh $(which ls)",
			"awk '{print $",
		}) {
			act = AttackActionHoneypotCheck
		} else if containsAny(attack.Contents, []string{
			"# uname",
			"# top",
			"cat /proc/cpuinfo",
			"cat /proc/mounts",
			"cat /proc/meminfo",
			"# id",
			"dd bs=52 count=1 if=",
			"lscpu | grep",
			"crontab -l",
		}) {
			act = AttackActionIdentification
		}
	}
	return act
}
