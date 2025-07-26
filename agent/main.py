#!/usr/bin/env python3
"""
Paperplay主程序 - 批量处理论文文件
完整的论文处理流程：PDF解析 -> 数据库存储 -> 向量存储
"""

from agents.paper_processing_agent import run_paper_processing_agent, process_single_paper
from agents.concept_extraction_agent import ConceptExtractionAgent
from utils import download_paper_list, PDFTextExtractor
import os
import json
from pathlib import Path

def get_paper_files(papers_dir: str = "papers") -> list:
    """获取论文目录中的所有PDF文件"""
    paper_files = []
    
    if not os.path.exists(papers_dir):
        print(f"❌ 论文目录不存在: {papers_dir}")
        return paper_files
    
    for file in os.listdir(papers_dir):
        if file.endswith(".pdf"):
            paper_path = os.path.join(papers_dir, file)
            paper_files.append(paper_path)
    
    print(f"📁 找到 {len(paper_files)} 个PDF文件")
    return paper_files

def download_predefined_papers() -> dict:
    """下载预定义的论文列表"""
    paper_list_for_aiagent = [
        "https://arxiv.org/abs/1706.03762",  # Attention Is All You Need
        "https://arxiv.org/abs/1810.04805",  # BERT
        "https://arxiv.org/abs/2005.14165",  # GPT-3
        "https://arxiv.org/abs/2201.11903",  # Chain-of-Thought Prompting
        "https://arxiv.org/abs/2210.03629",  # ReAct
        "https://arxiv.org/abs/2302.04761",  # Toolformer
        "https://arxiv.org/abs/2304.03442",  # Generative Agents: Interactive Simulacra of Human Behavior
        "https://arxiv.org/abs/2310.08560",  # MEMGPT
        "https://arxiv.org/abs/2304.11477",  # LLM+P
        "https://arxiv.org/abs/2303.03378"   # PaLM-E: An Embodied Multimodal Language Model
    ]
    
    print("📥 开始下载预定义论文列表...")
    result = download_paper_list(paper_list_for_aiagent)
    return result

def process_papers_batch(paper_files: list) -> dict:
    """批量处理论文文件"""
    results = {
        "total": len(paper_files),
        "success": 0,
        "failed": 0,
        "skipped": 0,
        "details": []
    }
    
    print(f"\n🚀 开始批量处理 {len(paper_files)} 篇论文...")
    print("=" * 60)
    
    for i, paper_path in enumerate(paper_files, 1):
        print(f"\n📄 [{i}/{len(paper_files)}] 处理论文: {os.path.basename(paper_path)}")
        print("-" * 40)
        
        # 检查是否已有解析结果
        md_path = paper_path + ".md"
        if os.path.exists(md_path):
            results["skipped"] += 1
            results["details"].append({
                "file": paper_path,
                "status": "skipped",
                "message": "已有解析结果，跳过处理"
            })
            print(f"⏭️ 已有解析结果，跳过: {os.path.basename(paper_path)}")
            continue
        
        try:
            # 使用增强版agent处理单个论文
            result = process_single_paper(paper_path)
            
            if result["status"] == "success":
                results["success"] += 1
                results["details"].append({
                    "file": paper_path,
                    "status": "success",
                    "message": "处理完成"
                })
                print(f"✅ 论文处理成功: {os.path.basename(paper_path)}")
            else:
                results["failed"] += 1
                results["details"].append({
                    "file": paper_path,
                    "status": "failed",
                    "message": result.get("message", "未知错误")
                })
                print(f"❌ 论文处理失败: {os.path.basename(paper_path)}")
                print(f"   错误信息: {result.get('message', '未知错误')}")
                
        except Exception as e:
            results["failed"] += 1
            results["details"].append({
                "file": paper_path,
                "status": "failed", 
                "message": str(e)
            })
            print(f"❌ 论文处理异常: {os.path.basename(paper_path)}")
            print(f"   异常信息: {str(e)}")
    
    return results

def process_concepts_batch(paper_files: list) -> dict:
    """批量处理论文概念提取"""
    results = {
        "total": len(paper_files),
        "success": 0,
        "failed": 0,
        "skipped": 0,
        "details": []
    }
    
    print(f"\n🧠 开始批量概念提取 {len(paper_files)} 篇论文...")
    print("=" * 60)
    
    # 初始化概念提取agent和PDF提取器
    concept_agent = ConceptExtractionAgent()
    pdf_extractor = PDFTextExtractor()
    
    for i, paper_path in enumerate(paper_files, 1):
        print(f"\n🔍 [{i}/{len(paper_files)}] 概念提取: {os.path.basename(paper_path)}")
        print("-" * 40)
        
        # 检查是否已有解析结果
        md_path = paper_path + ".md"
        if not os.path.exists(md_path):
            results["failed"] += 1
            results["details"].append({
                "file": paper_path,
                "status": "failed",
                "message": "缺少解析结果文件，请先进行论文处理"
            })
            print(f"❌ 缺少解析结果文件: {os.path.basename(paper_path)}")
            continue
        
        try:
            # 1. 提取arXiv ID用于检查概念数据库
            arxiv_id = pdf_extractor.extract_arxiv_id(paper_path)
            
            # 2. 检查概念数据库中是否已有该论文
            if arxiv_id:
                existing_paper = concept_agent.get_paper_concepts(arxiv_id)
                if existing_paper and len(existing_paper.get('concepts', [])) > 0:
                    results["skipped"] += 1
                    results["details"].append({
                        "file": paper_path,
                        "status": "skipped",
                        "message": f"概念已存在，跳过处理 (已有{len(existing_paper['concepts'])}个概念)",
                        "arxiv_id": arxiv_id,
                        "concepts_count": len(existing_paper['concepts'])
                    })
                    print(f"⏭️ 概念已存在，跳过: {os.path.basename(paper_path)}")
                    print(f"   ArXiv ID: {arxiv_id}")
                    print(f"   已有概念数: {len(existing_paper['concepts'])}")
                    continue
            
            # 3. 首先提取PDF内容
            extraction_result = pdf_extractor.extract_text_from_pdf(paper_path)
            
            if extraction_result["status"] != "success":
                results["failed"] += 1
                results["details"].append({
                    "file": paper_path,
                    "status": "failed",
                    "message": f"PDF提取失败: {extraction_result.get('message', '未知错误')}"
                })
                print(f"❌ PDF提取失败: {os.path.basename(paper_path)}")
                continue
            
            # 4. 进行概念提取
            paper_data = extraction_result["paper"]
            concept_result = concept_agent.process_paper_concepts(paper_data)
            
            if concept_result["status"] == "success":
                results["success"] += 1
                results["details"].append({
                    "file": paper_path,
                    "status": "success",
                    "message": concept_result["message"],
                    "concepts_count": len(concept_result["concepts"]),
                    "arxiv_id": concept_result["arxiv_id"]
                })
                print(f"✅ 概念提取成功: {os.path.basename(paper_path)}")
                print(f"   提取概念数: {len(concept_result['concepts'])}")
                print(f"   ArXiv ID: {concept_result['arxiv_id']}")
            else:
                results["failed"] += 1
                results["details"].append({
                    "file": paper_path,
                    "status": "failed",
                    "message": concept_result.get("message", "概念提取失败")
                })
                print(f"❌ 概念提取失败: {os.path.basename(paper_path)}")
                print(f"   错误信息: {concept_result.get('message', '未知错误')}")
                
        except Exception as e:
            results["failed"] += 1
            results["details"].append({
                "file": paper_path,
                "status": "failed",
                "message": str(e)
            })
            print(f"❌ 概念提取异常: {os.path.basename(paper_path)}")
            print(f"   异常信息: {str(e)}")
    
    return results

def show_final_stats():
    """显示最终统计信息"""
    print("\n📊 获取系统统计信息...")
    
    # 获取数据库统计
    db_result = run_paper_processing_agent("请获取数据库统计信息", thread_id="stats_db")
    
    # 获取向量库统计  
    vector_result = run_paper_processing_agent("请获取向量库统计信息", thread_id="stats_vector")
    
    # 获取概念数据库统计
    try:
        concept_agent = ConceptExtractionAgent()
        concept_stats = concept_agent.get_database_stats()
        
        print("\n🧠 概念数据库统计:")
        print(f"  • 总论文数: {concept_stats.get('total_papers', 0)}")
        print(f"  • 总概念数: {concept_stats.get('total_concepts', 0)}")
        print(f"  • 平均每篇论文概念数: {concept_stats.get('avg_concepts_per_paper', 0)}")
        
        if concept_stats.get('top_cited_papers'):
            print("\n📈 高引用论文:")
            for paper in concept_stats['top_cited_papers'][:3]:
                print(f"  • {paper['title'][:50]}... (引用数: {paper['citation_count']})")
                
    except Exception as e:
        print(f"⚠️ 获取概念数据库统计失败: {e}")
    
    print("\n🎯 系统状态总览完成！")


def paper_process(): 
    print("论文处理流程")
    print("完整流程：PDF解析 → 数据库存储 → 向量存储")
    print("=" * 60)
    
    # 0. 下载预定义论文列表
    download_result = download_predefined_papers()
    
    if download_result["success"] > 0 or download_result["exists"] > 0:
        print(f"✅ 论文下载完成！新下载: {download_result['success']}, 已存在: {download_result['exists']}")
    else:
        print("⚠️ 没有成功下载新论文，继续处理现有文件...")
    
    # 1. 获取所有PDF文件
    paper_files = get_paper_files()
    
    if not paper_files:
        print("⚠️ 未找到PDF文件，程序退出")
        return
    
    # 2. 显示待处理文件列表
    print("\n📋 待处理论文列表:")
    for i, paper_path in enumerate(paper_files, 1):
        print(f"  {i}. {os.path.basename(paper_path)}")
    
    # 3. 询问用户是否继续
    try:
        confirm = input(f"\n❓ 确认处理这 {len(paper_files)} 个文件？(y/N): ").strip().lower()
        if confirm not in ['y', 'yes', 'Y', '是']:
            print("👋 用户取消操作，程序退出")
            return
    except KeyboardInterrupt:
        print("\n👋 用户中断操作，程序退出")
        return
    
    # 4. 批量处理论文
    results = process_papers_batch(paper_files)
    
    # 6. 显示处理结果摘要
    print("\n" + "=" * 60)
    print("📊 批量处理结果摘要")
    print("=" * 60)
    print(f"✅ 成功处理: {results['success']} 篇")
    print(f"⏭️ 已跳过: {results['skipped']} 篇")
    print(f"❌ 处理失败: {results['failed']} 篇")
    print(f"📝 总计: {results['total']} 篇")
    
    # 显示跳过的文件详情
    if results['skipped'] > 0:
        print(f"\n⏭️ 跳过的文件详情:")
        for detail in results['details']:
            if detail['status'] == 'skipped':
                print(f"  - {os.path.basename(detail['file'])}: {detail['message']}")
    
    # 7. 显示失败的文件详情
    if results['failed'] > 0:
        print(f"\n❌ 失败文件详情:")
        for detail in results['details']:
            if detail['status'] == 'failed':
                print(f"  - {os.path.basename(detail['file'])}: {detail['message']}")
    
    # 8. 显示系统统计信息
    if results['success'] > 0:
        show_final_stats()
    
    
    print("\n🎉 论文批量处理完成！")
    print("💡 提示: 现在可以通过向量搜索功能查找相关论文了")

# def question_generate():
#         # 设计prompt
#         return []

def main():
    """主函数"""
    #论文处理
    paper_process()
    # #题目生成
    # question_generate()

if __name__ == "__main__":
    main()
    