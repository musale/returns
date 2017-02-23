#!/usr/bin/python2.7

import random
import urllib
import urllib2
from datetime import datetime, timedelta
from hashlib import md5

from faker import Faker

url = "http://callbacks.xsmsl.com/"
fake = Faker()


def get_status():
    return random.choice(['Success', 'Failed', 'Rejected'])


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


def get_reason(status):
    reason = 'subscriberNotExist'
    if status == 'Failed':
        return 'userUnreachable'
    return reason


def send_dlr(idx=None):
    status = get_status()
    payload = {
        'id': idx or md5(str(datetime.now())).hexdigest(),
        'status': status,
    }
    if status == 'Failed' or status == 'Rejected':
        payload['failureReason'] = get_reason(status)
    return urllib2.urlopen(url + 'at-dlrs', urllib.urlencode(payload)).read()


def send_inbox():
    payload = {
        'from': get_phone(), 'to': get_code(), 'text': get_message(),
        'date': str(datetime.now()), 'id': md5(str(datetime.now())).hexdigest()
    }
    return urllib2.urlopen('http://xsmsl.com/callbacks/inbox', urllib.urlencode(payload)).read()
    # return urllib2.urlopen(url + 'inbox', urllib.urlencode(payload)).read()


def pull_dlrs():
    import csv
    import MySQLdb as mdb

    host = 'localhost'
    user = 'kip'
    passw = 'kip@db'
    db = 'smsleopard'

    db = mdb.connect(
        host=host, user=user, passwd=passw, db=db)

    cur = db.cursor(mdb.cursors.DictCursor)

    sql = """
select api_id from bsms_smsrecipient where api_id is not null
"""
    cur.execute(sql)

    aids = []
    for aid in cur.fetchall():
        rid = aid['api_id']
        if len(rid) > 2:
            aids.append([rid])

    with open('dlr_reports.csv', 'w') as fp:
        a = csv.writer(fp, delimiter=',')
        a.writerows(aids)
    return 'Ready'


def push_dlrs():
    dlrs = []
    with open('/home/ekt/Desktop/dlr_reports.csv', 'r') as f:
        for x in f.readlines():
            dlrs.append(x.strip())
    for x in dlrs[20]:
        print send_dlr(x)
    return


def get_status_rm():
    return random.choice([
        'UNKNOWN', 'ACKED', 'ENROUTE', 'DELIVRD',
        'EXPIRED', 'DELETED', 'UNDELIV', 'ACCEPTED', 'REJECTD'
    ])


def send_rms_dlr(idx):
    s = random.randint(3, 100)
    payload = {
        'sStatus': get_status_rm(), 'reason': None,
        'sMessageId': idx or md5('hello').hexdigest(),
        'sSender': 'SMSLEOPARD', 'sMobileNo': '727372285',
        'dtDone': str(datetime.now())[:19],
        'dtSubmit': str(datetime.now() - timedelta(seconds=s))
    }
    return urllib2.urlopen(url + 'rm-dlrs', urllib.urlencode(payload)).read()


def push_rms_dlrs():
    dlrs = []
    with open('/home/ekt/Desktop/dlr_reports.csv', 'r') as f:
        for x in f.readlines():
            dlrs.append(x.strip())
    for x in dlrs[20:]:
        print send_rms_dlr(x)
    return


if __name__ == '__main__':
    # print send_inbox()
    for i in xrange(2200):
        print send_inbox()
    #     print send_dlr()
    # print push_dlrs()
    # print push_rms_dlrs()
    print "Done"
