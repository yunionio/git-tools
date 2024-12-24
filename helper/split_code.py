import os
import math

from IPython.core.application import shutil

# 源代码目录
SOURCE_DIR = "./cloudpods"
# 每个子目录的最大行数
MAX_LINES = 70000
# 输出目标根目录
OUTPUT_DIR = "split_code"

# 创建输出目录
if not os.path.exists(OUTPUT_DIR):
    os.makedirs(OUTPUT_DIR)

line_count = 0
part_cnt = 0

# 遍历源目录中的每个文件
for root, dirs, files in os.walk(SOURCE_DIR):
    for file_name in files:
        file_path = os.path.join(root, file_name)

        try:
            # 获取文件的行数
            with open(file_path, 'r', encoding='utf-8') as f:
                lines = f.readlines()
            line_count = line_count + len(lines)
        except Exception as e:
            print(f"read {file_path} exception: {e}, just copy it")

        if line_count >= MAX_LINES:
            part_cnt += 1
            line_count = 0
            print(f"生成 part {part_cnt}")

        src_file_path = file_path
        output_file_path = src_file_path.removeprefix(SOURCE_DIR)

        output_file = os.path.join(OUTPUT_DIR, f'{SOURCE_DIR}_{part_cnt}', f'./{output_file_path}')
        output_dir = os.path.dirname(output_file)
        # import  ipdb; ipdb.set_trace()
        if not os.path.exists(output_dir):
            os.makedirs(output_dir)
        shutil.copy(file_path, output_file)

# copy vendor
def copy_vendor(src_dir, part_cnt):
    vendor_dir = 'vendor'
    output_dir = os.path.join(OUTPUT_DIR, f'{SOURCE_DIR}_{part_cnt+1}', vendor_dir)
    print(f"拷贝 vendor 到 {output_dir}")
    shutil.copytree(os.path.join(src_dir, vendor_dir), output_dir)

copy_vendor('./cloudpods.bk', part_cnt)


print(f"代码分割完成，存储在 {OUTPUT_DIR} 中。")

