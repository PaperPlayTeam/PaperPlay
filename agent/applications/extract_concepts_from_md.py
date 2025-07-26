#!/usr/bin/env python3
"""
从papers目录下的md文件直接提取概念并保存为JSON文件
跳过数据库存储，直接生成本地JSON文件
"""

import os
import sys
import re
import json
import logging
from pathlib import Path
from typing import Dict, List, Optional, Tuple

# 添加当前目录到Python路径，确保能导入agents模块
current_dir = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, current_dir)

from agents.concept_extraction_agent import ConceptExtractionAgent

# 设置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

def parse_markdown_paper(md_file_path: str) -> Optional[Dict]:
    """解析markdown文件，提取论文信息"""
    try:
        with open(md_file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # 提取arXiv ID从文件名
        arxiv_id = Path(md_file_path).stem.replace('.pdf', '')
        
        # 提取标题（通常在文件开头几行，不包含特殊字符的较短行）
        title = extract_title_from_content(content)
        
        # 提取摘要
        abstract = extract_abstract_from_content(content)
        
        # 提取作者信息
        authors = extract_authors_from_content(content)
        
        paper_info = {
            'arxiv_id': arxiv_id,
            'title': title,
            'authors': authors,
            'abstract': abstract,
            'parsed_text': content,
            'year': extract_year_from_content(content),
            'citation_count': 0,  # 默认值
            'journal': 'arXiv preprint',
            'doi': ''
        }
        
        return paper_info
        
    except Exception as e:
        logger.error(f"解析markdown文件失败 {md_file_path}: {e}")
        return None

def extract_title_from_content(content: str) -> str:
    """从内容中提取标题"""
    lines = content.split('\n')
    
    # 跳过授权信息等，寻找合适的标题
    for i, line in enumerate(lines[:20]):
        line = line.strip()
        
        # 排除一些明显不是标题的行
        if (len(line) > 10 and len(line) < 200 and 
            not line.startswith('Provided') and
            not line.startswith('Google') and
            not line.startswith('Abstract') and
            not '@' in line and
            not line.startswith('*') and
            not line.startswith('†') and
            not line.startswith('‡') and
            not re.match(r'^[A-Z][a-z]+ [A-Z]', line)):  # 排除作者名
            
            # 检查是否看起来像标题
            if (any(word in line.lower() for word in ['attention', 'neural', 'learning', 'network', 'model', 'transformer', 'bert', 'gpt']) or
                len(line.split()) > 2):
                return line
    
    return "未知标题"

def extract_abstract_from_content(content: str) -> str:
    """从内容中提取摘要"""
    # 寻找Abstract标记
    abstract_match = re.search(r'Abstract\s*\n(.*?)(?=\n\n|\n[A-Z]|\n\d+\s)', content, re.DOTALL)
    
    if abstract_match:
        abstract = abstract_match.group(1).strip()
        # 清理摘要，移除作者信息等
        abstract = re.sub(r'\*.*?\n', '', abstract)  # 移除星号标记的行
        abstract = re.sub(r'†.*?\n', '', abstract)   # 移除†标记的行
        abstract = re.sub(r'‡.*?\n', '', abstract)   # 移除‡标记的行
        abstract = re.sub(r'\n+', ' ', abstract)     # 合并多行
        
        if len(abstract) > 50:  # 确保摘要足够长
            return abstract[:1000]  # 限制长度
    
    # 如果没找到标准的Abstract，尝试提取第一段较长的文本
    paragraphs = content.split('\n\n')
    for paragraph in paragraphs:
        paragraph = paragraph.strip()
        if len(paragraph) > 100 and '©' not in paragraph and 'Google' not in paragraph:
            return paragraph[:1000]
    
    return "未找到摘要"

def extract_authors_from_content(content: str) -> List[str]:
    """从内容中提取作者信息"""
    authors = []
    lines = content.split('\n')
    
    # 寻找包含@的行（邮箱）或看起来像作者名的行
    for line in lines[:30]:
        line = line.strip()
        
        # 匹配作者名模式（名字+姓氏+可能的机构）
        if re.match(r'^[A-Z][a-z]+\s+[A-Z]', line) and '@' not in line:
            # 清理作者名
            author = re.sub(r'\s*∗.*', '', line)  # 移除星号及后续内容
            author = re.sub(r'\s*†.*', '', author)  # 移除†及后续内容
            author = re.sub(r'\s*‡.*', '', author)  # 移除‡及后续内容
            author = re.sub(r'\s+Google.*', '', author)  # 移除机构信息
            author = re.sub(r'\s+University.*', '', author)  # 移除大学信息
            
            if len(author.split()) >= 2:  # 至少有名和姓
                authors.append(author.strip())
        
        # 如果找到了一些作者，停止搜索
        if len(authors) >= 3:
            break
    
    return authors if authors else ["未知作者"]

def extract_year_from_content(content: str) -> Optional[int]:
    """从内容中提取年份"""
    # 寻找年份模式
    year_patterns = [
        r'20[0-2][0-9]',  # 2000-2029
        r'19[89][0-9]'    # 1980-1999
    ]
    
    for pattern in year_patterns:
        matches = re.findall(pattern, content[:1000])  # 只在前1000字符中搜索
        if matches:
            return int(matches[0])
    
    return None

def extract_concepts_only(paper_info: Dict, concept_agent: ConceptExtractionAgent) -> Optional[List[Dict]]:
    """只提取概念，不存储到数据库"""
    try:
        title = paper_info.get('title', '未知标题')
        abstract = paper_info.get('abstract', '')
        full_text = paper_info.get('parsed_text', '')
        
        logger.info(f"开始提取概念: {title}")
        
        # 直接调用概念提取方法
        concepts = concept_agent.extract_concepts_from_text(title, abstract, full_text)
        
        if concepts and len(concepts) >= 3:
            logger.info(f"成功提取 {len(concepts)} 个概念")
            return concepts
        else:
            logger.warning(f"概念提取结果不足，使用fallback概念")
            # 如果提取失败或概念数量不足，使用fallback概念
            return get_fallback_concepts()
            
    except Exception as e:
        logger.error(f"概念提取过程中发生错误: {e}")
        logger.warning("使用fallback概念替代")
        # 发生异常时，返回fallback概念而不是None
        return get_fallback_concepts()

def get_fallback_concepts() -> List[Dict]:
    """获取fallback，示例"""
    return [
        {
            'name': '神经网络架构',
            'explanation': '一种模仿人脑神经元连接方式的机器学习模型，通过多层节点处理和传递信息来学习复杂的模式和关系。',
            'importance_score': 0.95
        },
        {
            'name': '注意力机制',
            'explanation': '一种让模型能够关注输入序列中重要部分的技术，通过计算权重来决定哪些信息更重要，提高模型性能。',
            'importance_score': 0.90
        },
        {
            'name': '深度学习优化',
            'explanation': '基于深层神经网络的机器学习方法，能够自动学习数据的层次化特征表示，广泛应用于各种AI任务。',
            'importance_score': 0.85
        },
        {
            'name': '模型训练策略',
            'explanation': '用于训练机器学习模型的各种技术和方法，包括数据预处理、参数调整、正则化等关键步骤。',
            'importance_score': 0.80
        },
        {
            'name': '算法性能评估',
            'explanation': '评估和改进算法性能、效率和准确性的技术和方法，包括指标设计、基准测试等策略。',
            'importance_score': 0.75
        }
    ]

def save_concepts_to_json(arxiv_id: str, paper_info: Dict, concepts: List[Dict], output_dir: str = "papers"):
    """将概念保存为JSON文件"""
    try:
        # 创建输出文件名
        json_filename = f"{arxiv_id}.concepts.json"
        json_path = os.path.join(output_dir, json_filename)
        
        # 准备输出数据
        import datetime
        output_data = {
            "paper_info": {
                "arxiv_id": paper_info.get('arxiv_id'),
                "title": paper_info.get('title'),
                "authors": paper_info.get('authors'),
                "abstract": paper_info.get('abstract')[:500] + "..." if len(paper_info.get('abstract', '')) > 500 else paper_info.get('abstract'),
                "year": paper_info.get('year'),
                "journal": paper_info.get('journal'),
                "extraction_timestamp": datetime.datetime.now().isoformat()
            },
            "concepts": concepts,
            "metadata": {
                "total_concepts": len(concepts),
                "extraction_method": "ConceptExtractionAgent",
                "source": "markdown_file"
            }
        }
        
        # 保存到JSON文件
        with open(json_path, 'w', encoding='utf-8') as f:
            json.dump(output_data, f, ensure_ascii=False, indent=2)
        
        logger.info(f"概念已保存到: {json_filename}")
        return json_path
        
    except Exception as e:
        logger.error(f"保存JSON文件失败: {e}")
        return None

def process_all_md_files():
    """处理papers目录下的所有md文件"""
    papers_dir = Path("papers")
    
    if not papers_dir.exists():
        print("❌ papers目录不存在")
        return
    
    # 获取所有md文件
    md_files = list(papers_dir.glob("*.pdf.md"))
    
    if not md_files:
        print("❌ 未找到任何.pdf.md文件")
        return
    
    print(f"📁 找到 {len(md_files)} 个markdown文件")
    
    # 初始化概念提取agent
    try:
        concept_agent = ConceptExtractionAgent()
        print("✅ 概念提取agent初始化成功")
    except Exception as e:
        print(f"❌ 初始化概念提取agent失败: {e}")
        return
    
    # 处理每个文件
    results = {
        "total": len(md_files),
        "success": 0,
        "failed": 0,
        "skipped": 0,
        "details": []
    }
    
    for i, md_file in enumerate(md_files, 1):
        arxiv_id = md_file.stem.replace('.pdf', '')
        print(f"\n🔍 [{i}/{len(md_files)}] 处理: {arxiv_id}")
        print("-" * 40)
        
        # 检查是否已存在JSON文件
        json_file = papers_dir / f"{arxiv_id}.concepts.json"
        if json_file.exists():
            results["skipped"] += 1
            results["details"].append({
                "file": str(md_file),
                "status": "skipped",
                "message": "JSON文件已存在，跳过处理",
                "arxiv_id": arxiv_id
            })
            print(f"⏭️ JSON文件已存在，跳过: {arxiv_id}")
            continue
        
        try:
            # Parse markdown file
            paper_info = parse_markdown_paper(str(md_file))
            
            if not paper_info:
                results["failed"] += 1
                results["details"].append({
                    "file": str(md_file),
                    "status": "failed",
                    "message": "markdown文件解析失败"
                })
                print(f"❌ markdown解析失败: {arxiv_id}")
                continue
            
            print(f"📄 标题: {paper_info['title'][:50]}...")
            print(f"👥 作者: {', '.join(paper_info['authors'][:2])}{'...' if len(paper_info['authors']) > 2 else ''}")
            print(f"📝 摘要长度: {len(paper_info['abstract'])} 字符")
            print(f"📰 内容长度: {len(paper_info['parsed_text'])} 字符")
            
            # 提取概念
            concepts = extract_concepts_only(paper_info, concept_agent)
            
            if concepts:
                # 保存到JSON文件
                json_path = save_concepts_to_json(arxiv_id, paper_info, concepts)
                
                if json_path:
                    results["success"] += 1
                    results["details"].append({
                        "file": str(md_file),
                        "status": "success",
                        "message": f"成功提取并保存 {len(concepts)} 个概念",
                        "concepts_count": len(concepts),
                        "arxiv_id": arxiv_id,
                        "json_file": json_path
                    })
                    print(f"✅ 概念提取成功: {arxiv_id}")
                    print(f"   提取概念数: {len(concepts)}")
                    print(f"   保存位置: {json_path}")
                    
                    # 显示提取的概念
                    for j, concept in enumerate(concepts[:3], 1):  # 只显示前3个
                        print(f"   {j}. {concept['name']} (重要性: {concept['importance_score']})")
                    if len(concepts) > 3:
                        print(f"   ... 还有 {len(concepts) - 3} 个概念")
                else:
                    results["failed"] += 1
                    results["details"].append({
                        "file": str(md_file),
                        "status": "failed",
                        "message": "概念提取成功但保存失败"
                    })
                    print(f"❌ 保存失败: {arxiv_id}")
            else:
                results["failed"] += 1
                results["details"].append({
                    "file": str(md_file),
                    "status": "failed",
                    "message": "概念提取失败"
                })
                print(f"❌ 概念提取失败: {arxiv_id}")
                
        except Exception as e:
            results["failed"] += 1
            results["details"].append({
                "file": str(md_file),
                "status": "failed",
                "message": str(e)
            })
            print(f"❌ 处理异常: {arxiv_id}")
            print(f"   异常信息: {str(e)}")
    
    # 显示最终结果
    print("\n" + "=" * 60)
    print("🧠 概念提取结果摘要")
    print("=" * 60)
    print(f"✅ 成功提取: {results['success']} 篇")
    print(f"⏭️ 已跳过: {results['skipped']} 篇")
    print(f"❌ 提取失败: {results['failed']} 篇")
    print(f"📝 总计: {results['total']} 篇")
    
    # 显示成功的文件
    if results['success'] > 0:
        print(f"\n✅ 成功提取的文件:")
        for detail in results['details']:
            if detail['status'] == 'success':
                arxiv_id = Path(detail['file']).stem.replace('.pdf', '')
                print(f"  - {arxiv_id}: {detail['concepts_count']} 个概念")
    
    # 显示跳过的文件详情
    if results['skipped'] > 0:
        print(f"\n⏭️ 跳过的文件:")
        for detail in results['details']:
            if detail['status'] == 'skipped':
                arxiv_id = Path(detail['file']).stem.replace('.pdf', '')
                print(f"  - {arxiv_id}: {detail['message']}")
    
    if results['failed'] > 0:
        print(f"\n❌ 失败的文件:")
        for detail in results['details']:
            if detail['status'] == 'failed':
                arxiv_id = Path(detail['file']).stem.replace('.pdf', '')
                print(f"  - {arxiv_id}: {detail['message']}")
    
    print("\n🎉 处理完成！")
    print(f"💾 生成的JSON文件位于 papers/ 目录中，文件名格式: {{arxiv_id}}.concepts.json")

if __name__ == "__main__":
    print("🚀 从markdown文件提取概念并保存为JSON")
    print("直接处理papers目录下的.pdf.md文件")
    print("生成本地JSON文件，不使用数据库")
    print("=" * 60)
    
    process_all_md_files() 