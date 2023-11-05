import re
import time
import random
import argparse
from pynput.keyboard import Controller
import pyperclip
import segno

parser = argparse.ArgumentParser()

rfid_badge = '@4324325553#'
qr_badge = '$4324325553#'

keyboard = Controller()  # Create the controller

def type_string_with_delay(string):
    delay = 2
    time.sleep(delay)  # Sleep for the amount of seconds generated
    for character in string:  # Loop over each character in the string
        keyboard.type(character)  # Type the character

parser.add_argument("-j", "--JIRA",
                    dest="jira_commits",
                    help="Jira commits")
parser.add_argument("-r", "--RFID",
                    dest="rfid_badge",
                    help="RFID badge")
parser.add_argument("-qr", "--QR",
                    dest="qr_badge",
                    help="QR badge")
parser.add_argument("-ng", "--NGROK",
                    dest="ngrok",
                    help="NGROK")
parser.add_argument("-st", "--STASH",
                    dest="stash",
                    help="STASH")
args = parser.parse_args()

if args.rfid_badge is not None:
    badge = f'@{args.rfid_badge}#'
    print(f'RFID badge: {badge}')
    type_string_with_delay(badge)
elif args.qr_badge is not None:
    badge = f'${args.qr_badge}#'
    print(f'QR badge: {badge}')
    type_string_with_delay(badge)
elif args.jira_commits is not None:
    textString = f'{args.jira_commits}'
    jiraBranchName = re.sub(r"[ !+@#$%^&*()':;>-]", "-", textString.lower().replace('\n', '-')).replace('suite', 'SUITE')
    jiraBranchName = re.sub(r'-+', '-', jiraBranchName)
    pyperclip.copy('git checkout -b ' + jiraBranchName)
    print('\n', jiraBranchName, '\n')
elif args.stash is not None:
    stash_text_raw = f'[stash]{args.stash}'
    stash_text = re.sub(r"[ !+@#$%^&*()':;>-]", "-", stash_text_raw.lower().replace('\n', '-')).replace('suite', 'SUITE').replace('[stash]', 'git stash push -m ')
    stash_text = re.sub(r'-+', '-', stash_text)
    pyperclip.copy(stash_text)
    print('\n', stash_text, '\n')
elif args.ngrok is not None:
    ngrok_url = f'segno --compact {args.ngrok}/i/source/ui-kiosk/index.html'
    type_string_with_delay(ngrok_url)
    pyperclip.copy(ngrok_url)
    print('\n')
# else:
#     badge = ''
#     print(f'Enter your badge {badge}')
#     badge = input()
#     type_string_with_delay(badge)
