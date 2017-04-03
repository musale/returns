#!/usr/bin/env python

from fabric.api import env, cd, run, sudo

env.use_ssh_config = True
env.hosts = ["mpesa"]
code_dir = "/home/ekt/go/src/github.com/etowett/returns/"
install_dir = "/apps/returns/"


def deploy():
    with cd(code_dir):
        run("git pull origin master")
        run("go build")
        run("go install")
    sudo("systemctl stop callbacks")
    with cd(install_dir):
        run("rm returns")
        run("cp /home/ekt/go/bin/returns .")
    sudo("systemctl start callbacks")
    return
