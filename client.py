#!/usr/bin/python2.7

import random
import urllib
import urllib2
from datetime import datetime
from hashlib import md5

from faker import Faker

url = "http://127.0.0.1:8017/"
fake = Faker()


def get_status():
    return random.choice(['Success', 'Failed'])


def get_phone():
    prefix = ['+2547', '+2557', '+2536', '+2116', '+2119', '+2567']
    net = [
        '34', '75', '16', '17', '18', '79', '20', '21', '22', '23', '64',
        '25', '96', '27', '55',
    ]
    return random.choice(prefix) + random.choice(net) + \
        str(random.randint(111111, 999999))


def get_code():
    return random.choice(['31390', '20880'])


def get_message():
    message = fake.sentence()
    if random.choice([True, False]):
        message = random.choice(['kip', 'marto']) + " " + message
    return message


def send_dlr(idx=None):
    payload = {
        'id': idx or md5(str(datetime.now())).hexdigest(),
        'status': get_status()
    }
    return urllib2.urlopen(url + 'at-dlrs', urllib.urlencode(payload)).read()


def send_inbox():
    payload = {
        'from': get_phone(), 'to': get_code(), 'text': get_message(),
        'date': str(datetime.now()), 'id': md5(str(datetime.now())).hexdigest()
    }
    return urllib2.urlopen(url + 'inbox', urllib.urlencode(payload)).read()


if __name__ == '__main__':
    # print send_inbox()
    for i in xrange(20):
        print send_dlr()
    print "Done"
