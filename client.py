#!/usr/bin/python2.7

import random
import urllib
import urllib2

from datetime import datetime
from hashlib import md5

url = "http://127.0.0.1:4147/"


def get_status():
    return random.choice(['Success', 'Failed'])


def send_dlr(idx=None):
    payload = urllib.urlencode({
        'id': idx or md5(str(datetime.now())).hexdigest(),
        'status': get_status()
    })
    return urllib2.urlopen(url + 'at-dlrs', payload).read()


if __name__ == '__main__':
    idx = '442c849e6ab4caf9c7cf3644e389a1e5'
    print send_dlr()
