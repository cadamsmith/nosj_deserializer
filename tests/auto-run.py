import os
import subprocess

################################################################################
# TODO: Put the name of your executable here.
################################################################################
PROG_NAME = '../src/nosj_deserializer'


def check_valid(input_path):
    res = subprocess.run([PROG_NAME, input_path], capture_output=True)

    if res.returncode != 0:
        return False

    if len(res.stderr) != 0:
        return False

    got = res.stdout
    with open(input_path.replace('.input', '.output'), 'rb') as handle:
        want = handle.read()
    if got != want:
        return False

    return True

def check_invalid(input_path):
    res = subprocess.run([PROG_NAME, input_path], capture_output=True)

    if res.returncode != 66:
        return False

    if len(res.stderr) == 0:
        return False

    if not res.stderr.decode('ascii').startswith('ERROR -- '):
        return False

    return True

def main():
    errors = []

    spec_valid_files = os.listdir('./from-spec/valid/')
    for valid_path in spec_valid_files:
        if not valid_path.endswith('.input'):
            continue

        if not check_valid('./from-spec/valid/'+valid_path):
            errors.append('incorrect handling of valid file: ./from-spec/valid/'+valid_path)
        else:
            print('OK --', './from-spec/valid/'+valid_path)

    spec_error_paths = os.listdir('./from-spec/error/')
    for error_path in spec_error_paths:
        if not check_invalid('./from-spec/error/'+error_path):
            errors.append('incorrect handling of error file: ./from-spec/error/'+error_path)
        else:
            print('OK --', './from-spec/error/'+error_path)

    lecture_valid_paths = os.listdir('./from-lecture/valid/')
    for valid_path in lecture_valid_paths:
        if not valid_path.endswith('.input'):
            continue

        if not check_valid('./from-lecture/valid/'+valid_path):
            errors.append('incorrect handling of valid file: ./from-lecture/valid/'+valid_path)
        else:
            print('OK --', './from-lecture/valid/'+valid_path)

    lecture_error_paths = os.listdir('./from-lecture/error/')
    for error_path in lecture_error_paths:
        if not check_invalid('./from-lecture/error/'+error_path):
            errors.append('incorrect handling of error file: ./from-lecture/error/'+error_path)
        else:
            print('OK --', './from-lecture/error/'+error_path)

    for error in errors:
        print('ERROR --', error)

main()
