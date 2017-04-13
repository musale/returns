#!/usr/bin/python2.7

import random
import urllib
import urllib2
from datetime import datetime, timedelta
from hashlib import md5

from faker import Faker

# url = "http://callbacks.smsleopard.com/"
url = "http://callhttp://callbacks.local/inbox/"
fake = Faker()


def get_status():
    return random.choice(['Success'])
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


def send_cache_req(idx=None, rid=None):
    payload = {
        'api_id': idx or md5(str(datetime.now())).hexdigest(),
        'recipient_id': rid or random.randint(1, 9999),
    }
    return urllib2.urlopen(url + 'cache-dlr', urllib.urlencode(payload)).read()


def send_inbox():
    payload = {
        'from': get_phone(), 'to': get_code(), 'text': get_message(),
        'date': str(datetime.now()), 'id': md5(str(datetime.now())).hexdigest()
    }
    return urllib2.urlopen(
        'http://xsmsl.com/callbacks/inbox', urllib.urlencode(payload)).read()
    # return urllib2.urlopen(url + 'inbox', urllib.urlencode(payload)).read()


def pull_dlrs():
    import csv
    import MySQLdb as mdb

    host = 'localhost'
    user = 'kip'
    passw = 'kip@db'
    db = 'xsmsl'

    db = mdb.connect(
        host=host, user=user, passwd=passw, db=db)

    cur = db.cursor(mdb.cursors.DictCursor)

    sql = """
select id, api_id from bsms_smsrecipient where api_id is not null
and time_sent > '2017-02-21 00:00:00'
"""
    cur.execute(sql)

    aids = []
    for aid in cur.fetchall():
        if len(aid['api_id']) > 0:
            aids.append([aid['id'], aid['api_id']])

    with open('dlr_reports.csv', 'w') as fp:
        a = csv.writer(fp, delimiter=',')
        a.writerows(aids)
    return 'file written'


def cache_dlrs():
    dlrs = []
    with open('dlr_reports.csv', 'r') as f:
        for ex in f.readlines():
            x = ex.split(",")
            dlrs.append((x[0].strip(), x[1].strip(),))
    for x in dlrs:
        print send_cache_req(x[1], x[0])
    return 'Done'


def push_dlrs():
    dlrs = []
    with open('dlr_reports.csv', 'r') as f:
        for x in f.readlines():
            dlrs.append(x.strip())
    for x in dlrs:
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
        # 'sStatus': get_status_rm(),
        'sStatus': 'DELIVRD',
        'sMessageId': idx or md5('hello').hexdigest(),
        'sSender': 'SMSLEOPARD', 'sMobileNo': '727372285',
        'dtDone': str(datetime.now())[:19],
        'dtSubmit': str(datetime.now() - timedelta(seconds=s))
    }
    return urllib2.urlopen(url + 'rm-dlrs', urllib.urlencode(payload)).read()


def push_all_dlrs():
    dlrs = []
    with open('dlr_reports.csv', 'r') as f:
        for x in f.readlines():
            dlrs.append(x.strip())
    for x in dlrs:
        rid, aid = x.split(",")
        if len(aid) == 36:
            print send_rms_dlr(aid)
        else:
            print send_dlr(aid)
    return


def send_optout():
    payload = {
        'phoneNumber': get_phone(), 'senderId': 'SMSLEOPARD'
    }
    return urllib2.urlopen(url + 'optout', urllib.urlencode(payload)).read()


if __name__ == '__main__':

    # print pull_dlrs()
    # print cache_dlrs()

    # print push_all_dlrs()

    print send_rms_dlr('10f61d55-f863-45f6-8cfd-ea862abd102a')
    print send_dlr('5a2637fc-4b0e-48c0-be2b-1dab0ddcd2dd')

    # for i in xrange(20):
    #     print send_optout()
