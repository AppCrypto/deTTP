from __future__ import print_function
import os,sys
base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.append(base_dir)

from .curve import *


_hash_points_and_message = lambda a, b, m: hashsn(hashpn(a, b), m)

import hashlib
    



def hash(str):
    x = hashlib.sha256()
    x.update(str.encode())
    return x.hexdigest()


def schnorr_create(secret, message, id=None, point=None):
	assert isinstance(secret, long)
	assert isinstance(message, long)
	message+=int(hash(id)[:16],16)
	# print(message)
	xG = multiply(point, secret) if point else sbmul(secret)
	k = hashsn(message, secret)
	kG = multiply(point, k) if point else sbmul(k)
	e = hashs(xG[0].n, xG[1].n, kG[0].n, kG[1].n, message)
	s = submodn(k, mulmodn(secret, e))
	return [list(xG), s, e, message]


def schnorr_calc(xG, s, e, message, point=None):
	assert isinstance(s, long)
	assert isinstance(e, long)
	assert isinstance(message, long)
	sG = multiply(point, s) if point else sbmul(s)
	kG = add(sG, multiply(xG, e))
	return hashs(xG[0].n, xG[1].n, kG[0].n, kG[1].n, message)


def schnorr_verify(xG, s, e, message, id=None, point=None):
	message+=int(hash(id)[:16],16)
	return e == schnorr_calc(xG, s, e, message, point)


if __name__ == "__main__":
	s = 19977808579986318922850133509558564821349392755821541651519240729619349670944
	m = 19996069338995852671689530047675557654938145690856663988250996769054266469975
	proof = schnorr_create(s, m)
	assert proof[1] == 9937528682437333073292374920792423444152291976168124823244260606973530841357
	assert proof[2] == 62556699762868942562895201798238094653401696340984411017785245503967199042244
	print(schnorr_verify(*proof))
