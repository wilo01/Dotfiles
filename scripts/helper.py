import re
import time
import random
import argparse
import pyshorteners
from pynput.keyboard import Controller
import pyperclip
import segno
import shutil
import os

parser = argparse.ArgumentParser()
type_tiny = pyshorteners.Shortener()

rfid_badge = '@4324325553#'
qr_badge = '$4324325553#'

keyboard = Controller()  # Create the controller

def create_temp_dir(path):
    os.makedirs(path, exist_ok=True)
    return path

def copy_and_override_files(source1, source2, temp_dir):
    file_names1 = os.listdir(source1)
    file_names2 = os.listdir(source2)

    # Copy files from the first source directory to the temporary directory
    for file_name in file_names1:
        shutil.copy(os.path.join(source1, file_name), temp_dir)

    # Copy and overwrite files from the second source directory to the temporary directory
    for file_name in file_names2:
        shutil.copy(os.path.join(source2, file_name), temp_dir)

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
parser.add_argument("-vsc", "--VSC",
                    dest="vsc",
                    help="VSC")
parser.add_argument("-st", "--STASH",
                    dest="stash",
                    help="STASH")
parser.add_argument("-rt", "--RT",
                    dest="rt",
                    help="RT")
parser.add_argument("-po", "--PORT",
                    dest="port",
                    help="PORT")
parser.add_argument("-au", "--AUTH",
                    dest="auth",
                    help="AUTH")
parser.add_argument("-ap", "--APEX",
                    dest="apex",
                    help="APEX")
args = parser.parse_args()

if args.rfid_badge is not None:
    badge = f'@{args.rfid_badge}#'
    print(f'RFID badge: {badge}')
    type_string_with_delay(badge)
elif args.qr_badge is not None:
    badge = f'${args.qr_badge}#'
    print(f'QR badge: {badge}')
elif args.jira_commits is not None:
    textString = f'{args.jira_commits}'
    jiraBranchName = re.sub(r"[ !+@#$%^&*(),_.'/:;>-]", "-", textString.lower().replace('\n', '-')).replace('suite', 'SUITE').replace('vis-', 'VIS-').replace('bookr', 'BOOKR')
    jiraBranchName = re.sub(r'-+', '-', jiraBranchName)
    pyperclip.copy('git checkout -b ' + jiraBranchName)
    print('\n', jiraBranchName, '\n')
elif args.stash is not None:
    stash_text_raw = f'[stash]{args.stash}'
    stash_text = re.sub(r"[ !+@#$%^&*(),_.'/:;>-]", "-", stash_text_raw.lower().replace('\n', '-')).replace('suite', 'SUITE').replace('[stash]', 'git stash push -u -m ')
    stash_text = re.sub(r'-+', '-', stash_text)
    pyperclip.copy(stash_text)
    print('\n', stash_text, '\n')
elif args.ngrok is not None:
    ngrok_url = f'{args.ngrok}/i/source/ui-kiosk/index.html'
    print('\n' + ngrok_url + '\n')
    ngrok_short_url = type_tiny.tinyurl.short(ngrok_url)
    qr_code_url = f'segno --compact {ngrok_short_url}'
    pyperclip.copy(qr_code_url)
    print('\n' + qr_code_url + '\n')
elif args.vsc is not None:
    vsc_url = f'{args.vsc}i/source/ui-kiosk/index.html'
    print('\n' + vsc_url + '\n')
    vsc_short_url = type_tiny.tinyurl.short(vsc_url)
    qr_code_url = f'segno --compact {vsc_short_url}'
    pyperclip.copy(qr_code_url)
    print('\n' + qr_code_url + '\n')
elif args.rt is not None:
    path_to_temp_dir = f'~/Dev/branch-opener2/{args.rt}'
    source_dir1 = '~/Dev/branch-opener2/branches/safe/source/server/rt'
    source_dir2 = '~/Dev/branch-opener2/branches/safe/source/server/rtsp'
    temp_dir = create_temp_dir(path_to_temp_dir)
    copy_and_override_files(source_dir1, source_dir2, temp_dir)
    print('\n')
elif args.apex and args.port and args.auth is not None:
    apex_upload = f'curl -v --request POST --url http://127.0.0.1:8090/apex/_/resources.zip --header "Authorization: Basic {args.auth}" --form name==file --form "filename=@{args.apex}.zip;type=application/zip" | > {args.apex}.html'
    pyperclip.copy(apex_upload)
    print('\n' + apex_upload + '\n')
# else:
#     badge = ''
#     print(f'Enter your badge {badge}')
#     badge = input()
#     type_string_with_delay(badge)
