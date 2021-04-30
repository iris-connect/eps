import datetime
import os

script_dir = os.path.dirname(os.path.abspath(__file__))
source_dirs = [os.path.dirname(script_dir)]
context = {}

def read_copyright():
    with open(os.path.join(script_dir,'copyright_text.txt')) as input_file:
        return input_file.read().strip()

def format_copyright(text):
    c = context.copy()
    c['year'] = datetime.datetime.now().year
    return "\n".join([("// "+ t).strip() for t in text.format(**c).split("\n")])

def process_file(path, copyright_notice):
    with open(path) as input:
        content = input.read()
    lines = content.split("\n")
    for line in lines:
        if line.strip() == "// no-copyright-header":
            return
    for i, line in enumerate(lines):
        if not line.startswith('//'):
            break
    if lines[i].strip() == '':
        i+=1
    lines = lines[i:]
    new_content = copyright_notice+"\n\n"+"\n".join(lines)
    print("Writing {}".format(path))
    with open(path, 'w') as output:
        output.write(new_content)
    

def enumerate_files(dir, extensions=['.go'], exclude=set([])):
    for file in os.listdir(dir):
        path = os.path.join(dir, file)
        if file.startswith('.'):
            continue
        if os.path.isdir(path) and file not in exclude:
            for path in enumerate_files(path):
                yield path
        else:
            for extension in extensions:
                if file.endswith(extension):
                    yield path
if __name__ == '__main__':
    copyright_notice = format_copyright(read_copyright())
    for source_dir in source_dirs:
        for path in enumerate_files(source_dir):
            process_file(path, copyright_notice)
