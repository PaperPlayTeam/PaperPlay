#!/usr/bin/env python3
"""
论文信息存储脚本
从PDF文件或已解析的Markdown文件中提取论文信息并存储到paperplay.db数据库

使用方法:
python applications/store_papers_to_db.py
"""

import os
import sys
import json
import logging
from pathlib import Path
from typing import Dict, List, Optional
import datetime

# 添加项目根目录到Python路径
current_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.dirname(current_dir)
sys.path.insert(0, project_root)

# 导入项目模块
from utils.database_manager import DatabaseManager
from utils.pdf_text_extractor import PDFTextExtractor
from dotenv import load_dotenv

# 加载环境变量
env_path = os.path.join(project_root, '.env')
if os.path.exists(env_path):
    load_dotenv(env_path)
    print("✅ 已加载.env文件中的环境变量")
else:
    print("⚠️ 未找到.env文件，请确保环境变量已正确设置")

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class PaperStoreManager:
    """论文信息存储管理器"""
    
    def __init__(self, db_path: str = None):
        """初始化存储管理器"""
        if db_path is None:
            db_path = os.path.join(project_root, 'sqlite', 'paperplay.db')
        
        self.db_manager = DatabaseManager(db_path)
        self.pdf_extractor = PDFTextExtractor()
        self.logger = logger
        
        self.logger.info(f"论文存储管理器初始化: {db_path}")
        
    def extract_paper_info_from_pdf(self, pdf_path: str) -> Optional[Dict]:
        """从PDF文件提取论文信息"""
        try:
            result = self.pdf_extractor.extract_text_from_pdf(pdf_path)
            
            if result["status"] == "success":
                return result["paper"]
            else:
                self.logger.error(f"PDF提取失败: {result.get('message', '未知错误')}")
                return None
                
        except Exception as e:
            self.logger.error(f"PDF信息提取异常: {e}")
            return None
    
    def extract_paper_info_from_md(self, md_path: str) -> Optional[Dict]:
        """从Markdown文件提取论文信息"""
        try:
            with open(md_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            # 解析markdown内容，提取论文信息
            paper_info = self.parse_markdown_paper_info(content)
            
            # 从文件名获取arxiv_id
            filename = Path(md_path).stem
            if filename.endswith('.pdf'):
                arxiv_id = filename.replace('.pdf', '')
            else:
                arxiv_id = filename
                
            paper_info['arxiv_id'] = arxiv_id
            
            return paper_info
            
        except Exception as e:
            self.logger.error(f"Markdown信息提取异常: {e}")
            return None
    
    def parse_markdown_paper_info(self, content: str) -> Dict:
        """解析Markdown内容，提取论文基本信息"""
        paper_info = {
            'title': '未知标题',
            'authors': [],
            'abstract': '',
            'year': datetime.datetime.now().year,
            'journal': '',
            'parsed_text': content[:2000]  # 保留前2000字符作为内容预览
        }
        
        lines = content.split('\n')
        current_section = None
        
        for line in lines[:50]:  # 只检查前50行来提取元数据
            line = line.strip()
            
            # 提取标题（通常在开头的大标题）
            if line.startswith('# ') and paper_info['title'] == '未知标题':
                paper_info['title'] = line[2:].strip()
            
            # 提取作者信息
            if any(keyword in line.lower() for keyword in ['author', '作者', 'by ']):
                if ',' in line or ';' in line:
                    # 尝试解析作者列表
                    authors_text = line.split('author')[1] if 'author' in line.lower() else line
                    authors_text = authors_text.strip('s:：').strip()
                    if ',' in authors_text:
                        paper_info['authors'] = [author.strip() for author in authors_text.split(',')]
                    elif ';' in authors_text:
                        paper_info['authors'] = [author.strip() for author in authors_text.split(';')]
                    else:
                        paper_info['authors'] = [authors_text.strip()]
            
            # 提取年份
            if 'year' in line.lower() or '年' in line:
                import re
                years = re.findall(r'\b(19|20)\d{2}\b', line)
                if years:
                    paper_info['year'] = int(years[0])
            
            # 提取摘要部分
            if line.lower().startswith('abstract') or line.startswith('摘要'):
                current_section = 'abstract'
                continue
            elif current_section == 'abstract' and line:
                if line.startswith('#') or line.lower().startswith('introduction'):
                    current_section = None
                else:
                    paper_info['abstract'] += line + ' '
        
        # 清理摘要
        paper_info['abstract'] = paper_info['abstract'].strip()[:1000]  # 限制长度
        
        return paper_info
    
    def store_paper_to_db(self, paper_info: Dict) -> Optional[str]:
        """将论文信息存储到数据库"""
        try:
            # 准备数据库字段
            title = paper_info.get('title', '未知标题')[:500]  # 限制长度
            
            # 处理作者信息
            authors = paper_info.get('authors', [])
            if isinstance(authors, list):
                paper_author = '; '.join(authors)[:500]  # 限制长度
            else:
                paper_author = str(authors)[:500]
            
            # 处理发表年月
            year = paper_info.get('year', datetime.datetime.now().year)
            paper_pub_ym = str(year)
            
            # 插入到数据库
            paper_id = self.db_manager.insert_paper(
                title=title,
                paper_author=paper_author,
                paper_pub_ym=paper_pub_ym,
                paper_citation_count="0"  # 默认引用次数为0
            )
            
            self.logger.info(f"成功存储论文: {title[:50]}... (ID: {paper_id})")
            return paper_id
            
        except Exception as e:
            self.logger.error(f"存储论文到数据库失败: {e}")
            return None
    
    def check_paper_exists(self, title: str) -> bool:
        """检查论文是否已存在于数据库中"""
        try:
            # 这里可以实现检查逻辑，暂时返回False
            # 实际实现需要在DatabaseManager中添加相应方法
            return False
        except Exception as e:
            self.logger.error(f"检查论文存在性失败: {e}")
            return False
    
    def process_single_file(self, file_path: str) -> Dict:
        """处理单个文件"""
        filename = os.path.basename(file_path)
        
        try:
            # 根据文件类型选择处理方法
            if file_path.endswith('.pdf'):
                paper_info = self.extract_paper_info_from_pdf(file_path)
            elif file_path.endswith('.md'):
                paper_info = self.extract_paper_info_from_md(file_path)
            else:
                return {
                    "status": "skipped",
                    "message": f"不支持的文件类型: {filename}"
                }
            
            if not paper_info:
                return {
                    "status": "failed",
                    "message": "论文信息提取失败"
                }
            
            # 检查是否已存在（可选）
            title = paper_info.get('title', '')
            if self.check_paper_exists(title):
                return {
                    "status": "skipped", 
                    "message": "论文已存在于数据库中"
                }
            
            # 存储到数据库
            paper_id = self.store_paper_to_db(paper_info)
            
            if paper_id:
                return {
                    "status": "success",
                    "paper_id": paper_id,
                    "title": paper_info.get('title', ''),
                    "authors": paper_info.get('authors', []),
                    "year": paper_info.get('year', '')
                }
            else:
                return {
                    "status": "failed",
                    "message": "数据库存储失败"
                }
                
        except Exception as e:
            return {
                "status": "failed",
                "message": f"处理异常: {str(e)}"
            }

def find_paper_files(directory: str = "papers") -> List[str]:
    """查找papers目录下的论文文件"""
    paper_files = []
    papers_dir = Path(directory)
    
    if not papers_dir.exists():
        print(f"❌ 目录不存在: {directory}")
        return paper_files
    
    # 查找PDF文件
    pdf_files = list(papers_dir.glob("*.pdf"))
    
    # 查找Markdown文件
    md_files = list(papers_dir.glob("*.pdf.md"))
    
    # 优先使用PDF文件，如果没有PDF则使用MD文件
    processed_arxiv_ids = set()
    
    for pdf_file in pdf_files:
        arxiv_id = pdf_file.stem
        paper_files.append(str(pdf_file))
        processed_arxiv_ids.add(arxiv_id)
    
    # 添加没有对应PDF的MD文件
    for md_file in md_files:
        arxiv_id = md_file.stem.replace('.pdf', '')
        if arxiv_id not in processed_arxiv_ids:
            paper_files.append(str(md_file))
    
    return sorted(paper_files)

def process_all_papers():
    """处理所有论文文件"""
    print("🚀 论文信息存储脚本")
    print("从论文文件中提取信息并存储到paperplay.db数据库")
    print("=" * 60)
    
    # 查找论文文件
    paper_files = find_paper_files()
    
    if not paper_files:
        print("❌ 未找到任何论文文件")
        return
    
    print(f"📁 找到 {len(paper_files)} 个论文文件")
    
    # 初始化存储管理器
    try:
        store_manager = PaperStoreManager()
        print("✅ 存储管理器初始化成功")
    except Exception as e:
        print(f"❌ 存储管理器初始化失败: {e}")
        return
    
    # 处理结果统计
    results = {
        "total_files": len(paper_files),
        "success_count": 0,
        "failed_count": 0,
        "skipped_count": 0,
        "details": []
    }
    
    # 批量处理
    for i, paper_file in enumerate(paper_files, 1):
        filename = os.path.basename(paper_file)
        print(f"\n🔍 [{i}/{len(paper_files)}] 处理: {filename}")
        print("-" * 50)
        
        result = store_manager.process_single_file(paper_file)
        results["details"].append(result)
        
        if result["status"] == "success":
            results["success_count"] += 1
            print(f"  ✅ 存储成功")
            print(f"     论文ID: {result['paper_id']}")
            print(f"     标题: {result['title'][:50]}...")
            print(f"     作者: {', '.join(result['authors'][:3])}{'...' if len(result['authors']) > 3 else ''}")
            print(f"     年份: {result['year']}")
            
        elif result["status"] == "skipped":
            results["skipped_count"] += 1
            print(f"  ⏭️ 已跳过: {result['message']}")
            
        else:  # failed
            results["failed_count"] += 1
            print(f"  ❌ 处理失败: {result['message']}")
    
    # 显示最终结果
    print("\n" + "=" * 60)
    print("🎯 处理结果摘要")
    print("=" * 60)
    print(f"📁 文件处理: {results['success_count']}成功 / {results['failed_count']}失败 / {results['skipped_count']}跳过 / {results['total_files']}总计")
    print(f"📊 成功率: {results['success_count']/results['total_files']*100:.1f}%" if results['total_files'] > 0 else "📊 成功率: 0%")
    
    if results['success_count'] > 0:
        print(f"\n✅ 成功存储的论文:")
        for detail in results['details']:
            if detail['status'] == 'success':
                print(f"  - {detail['title'][:50]}... (ID: {detail['paper_id']})")
    
    print(f"\n💾 所有论文信息已存储到数据库: sqlite/paperplay.db")
    print("🎉 处理完成！")

if __name__ == "__main__":
    print("🚀 论文信息存储脚本")
    print("从论文文件中提取信息并存储到paperplay.db数据库")
    print("=" * 60)
    
    process_all_papers() 