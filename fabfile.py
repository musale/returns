
from fabric.colors import green, red
from fabric.contrib.files import exists
from fabric.api import env, cd, run, sudo
from fabric.contrib.project import rsync_project

env.use_ssh_config = True
env.hosts = ["web"]
install_dir = "/apps/returns"
home_dir = "/home/focus"
local_dir = "/home/ekt/go/src/github.com/etowett/"
live_dir = "%s/go/src/github.com/etowett/" % home_dir
user = "focus"


def stage():
    env.hosts = ["sms"]
    return


def deploy():
    with cd("%returns" % live_dir):
        print(green("Pull changes from bitbucket"))
        run("git pull origin master")
        print(green("get dependencies if any"))
        run("go get")
        print(green("build"))
        run("go build")
        print(green("install new"))
        run("go install")
    print(green("populating redis"))
    run("go run %sreturns/scripts/cache-codes.go")
    print(red("stop returns application"))
    stop_returns()
    with cd(install_dir):
        if exists("returns"):
            print(red("remove old returns"))
            run("rm returns")
        print(green("copy new returns"))
        run("cp %s/go/bin/returns ." % home_dir)
    print(green("start returns application"))
    restart_returns()
    return


def xdeploy():
    rsync_project(
        live_dir, local_dir='%sreturns' % local_dir,
        exclude=['*.pyc', '.git*'], delete=True
    )
    with cd('%sreturns' % live_dir):
        print(green("get dependencies if any"))
        run('go get')
        print(green("build"))
        run('go build')
        print(green("install new"))
        run('go install')
    print(green("populating redis"))
    run("go run %sreturns/scripts/cache-codes.go" % live_dir)
    with cd(install_dir):
        if exists("returns"):
            print(red("remove old returns"))
            run("rm returns")
        print(green("copy new returns"))
        run("cp /home/focus/go/bin/returns .")
    print(green("start service"))
    restart_returns()
    return


def setup():
    sudo("yum -y install go git nginx")
    if not exists("/home/focus/go"):
        run("mkdir /home/focus/go")
        run("echo \"export GOPATH=$HOME/go\" >> /home/focus/.bashrc")
    if not exists('%smoniyee' % live_dir):
        rsync_project(
            live_dir, local_dir='%sreturns' % local_dir,
            exclude=['*.pyc', '.git*'], delete=True
        )
    with cd('%sreturns' % live_dir):
        run('go get')
        run('go build')
        run('go install')
    if not exists("/apps"):
        sudo("mkdir /apps")
        sudo("chown %s:%s /apps" % (user, user,))
    with cd("/apps"):
        if not exists("returns"):
            run("mkdir returns")
        with cd("returns"):
            run("cp %sreturns/env.sample .env" % (live_dir,))
            run("cp /home/focus/go/bin/returns .")
    sudo(
        "cp %sreturns/config/callbacks.service "
        "/etc/systemd/system/callbacks.service" % (live_dir,)
    )
    sudo(
        "cp %sreturns/config/callbacks.conf "
        "/etc/nginx/conf.d/" % (live_dir,)
    )
    with cd("/var/log/"):
        if not exists("returns"):
            sudo("mkdir returns")
            sudo("chown %s:%s returns" % (user, user,))
        with cd("returns"):
            run("touch returns.log")
    restart_returns()
    return


def stop_returns():
    sudo("systemctl stop callbacks")
    return


def restart_returns():
    sudo('systemctl restart callbacks')
    return
