#!/usr/bin/env python3
"""
论文下载工具 - 从arXiv链接批量下载PDF文件到papers目录
"""

import requests
import os
import re
from pathlib import Path
from typing import List, Dict
import logging

class PaperDownloader:
    """论文下载器 - 专门用于下载arXiv论文到papers目录"""
    
    def __init__(self, papers_dir: str = "papers"):
        self.papers_dir = papers_dir
        self.logger = logging.getLogger(__name__)
        
        # 确保papers目录存在
        os.makedirs(papers_dir, exist_ok=True)
        
    def extract_arxiv_id_from_url(self, arxiv_url: str) -> str:
        """从arXiv URL中提取完整的arXiv ID（包含版本号）"""
        # 匹配arXiv URL格式: https://arxiv.org/abs/1706.03762 或 https://arxiv.org/abs/1706.03762v2
        patterns = [
            r'https?://arxiv\.org/abs/([0-9]{4}\.[0-9]{4,5}(?:v[0-9]+)?)',
            r'https?://arxiv\.org/pdf/([0-9]{4}\.[0-9]{4,5}(?:v[0-9]+)?)',
            r'([0-9]{4}\.[0-9]{4,5}(?:v[0-9]+)?)'
        ]
        
        for pattern in patterns:
            match = re.search(pattern, arxiv_url)
            if match:
                return match.group(1)
        
        raise ValueError(f"无法从URL中提取arXiv ID: {arxiv_url}")
    
    def extract_base_arxiv_id(self, arxiv_id: str) -> str:
        """从完整arXiv ID中提取基础ID（去除版本号）"""
        # 去除版本号，例如 1706.03762v2 -> 1706.03762
        return re.sub(r'v[0-9]+$', '', arxiv_id)
    
    def extract_version_number(self, arxiv_id: str) -> int:
        """从arXiv ID中提取版本号，如果没有版本号则返回1"""
        version_match = re.search(r'v([0-9]+)$', arxiv_id)
        return int(version_match.group(1)) if version_match else 1
    
    def find_existing_versions(self, base_arxiv_id: str) -> List[Dict[str, any]]:
        """查找已存在的该论文的所有版本"""
        existing_versions = []
        
        # 检查papers目录中的文件
        for filename in os.listdir(self.papers_dir):
            if filename.startswith(base_arxiv_id) and filename.endswith('.pdf'):
                file_path = os.path.join(self.papers_dir, filename)
                # 提取版本号
                version_match = re.search(f'{re.escape(base_arxiv_id)}(?:v([0-9]+))?\\.pdf$', filename)
                version = int(version_match.group(1)) if version_match and version_match.group(1) else 1
                
                existing_versions.append({
                    'filename': filename,
                    'path': file_path,
                    'version': version,
                    'size': os.path.getsize(file_path)
                })
        
        return sorted(existing_versions, key=lambda x: x['version'], reverse=True)
    
    def build_pdf_url(self, arxiv_id: str) -> str:
        """根据arXiv ID构建PDF下载链接"""
        return f"https://arxiv.org/pdf/{arxiv_id}.pdf"
    
    def get_paper_filename(self, arxiv_id: str, use_base_id: bool = True) -> str:
        """根据arXiv ID生成文件名"""
        if use_base_id:
            # 使用基础ID（去除版本号）作为文件名
            base_id = self.extract_base_arxiv_id(arxiv_id)
            return f"{base_id}.pdf"
        else:
            # 使用完整ID（包含版本号）作为文件名
            return f"{arxiv_id}.pdf"
    
    def download_single_paper(self, arxiv_url: str, force_latest: bool = True) -> Dict[str, any]:
        """下载单个论文，智能处理版本"""
        try:
            # 1. 提取完整arXiv ID（可能包含版本号）
            full_arxiv_id = self.extract_arxiv_id_from_url(arxiv_url)
            base_arxiv_id = self.extract_base_arxiv_id(full_arxiv_id)
            current_version = self.extract_version_number(full_arxiv_id)
            
            # 2. 检查是否已有该论文的其他版本
            existing_versions = self.find_existing_versions(base_arxiv_id)
            
            if existing_versions:
                latest_existing = existing_versions[0]  # 已按版本号降序排序
                existing_version = latest_existing['version']
                
                # 如果已有更新或相同版本，跳过下载
                if existing_version >= current_version:
                    return {
                        "status": "exists",
                        "arxiv_id": base_arxiv_id,
                        "full_arxiv_id": full_arxiv_id,
                        "filename": latest_existing['filename'],
                        "file_path": latest_existing['path'],
                        "size": latest_existing['size'],
                        "existing_version": existing_version,
                        "requested_version": current_version,
                        "message": f"已存在更新版本 v{existing_version}（请求版本 v{current_version}）: {latest_existing['filename']}"
                    }
                
                # 如果请求的是更新版本，删除旧版本
                if force_latest and current_version > existing_version:
                    for old_version in existing_versions:
                        old_file_path = old_version['path']
                        old_md_path = old_file_path + '.md'
                        
                        self.logger.info(f"删除旧版本 v{old_version['version']}: {old_version['filename']}")
                        os.remove(old_file_path)
                        
                        # 同时删除对应的md文件
                        if os.path.exists(old_md_path):
                            os.remove(old_md_path)
                            self.logger.info(f"删除旧版本解析文件: {old_version['filename']}.md")
            
            # 3. 生成新文件路径（使用基础ID）
            filename = self.get_paper_filename(full_arxiv_id, use_base_id=True)
            file_path = os.path.join(self.papers_dir, filename)
            
            # 4. 构建PDF下载链接
            pdf_url = self.build_pdf_url(full_arxiv_id)
            
            # 5. 下载文件
            self.logger.info(f"下载论文: {full_arxiv_id} -> {filename}")
            
            headers = {
                'User-Agent': 'Mozilla/5.0 (Linux; x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
            }
            
            response = requests.get(pdf_url, headers=headers, timeout=30, stream=True)
            response.raise_for_status()
            
            # 6. 保存文件
            with open(file_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            
            file_size = os.path.getsize(file_path)
            
            version_info = f" (v{current_version})" if current_version > 1 else ""
            
            return {
                "status": "success",
                "arxiv_id": base_arxiv_id,
                "full_arxiv_id": full_arxiv_id,
                "filename": filename,
                "file_path": file_path,
                "size": file_size,
                "version": current_version,
                "message": f"下载成功{version_info}: {filename}"
            }
            
        except Exception as e:
            error_msg = f"下载失败: {str(e)}"
            self.logger.error(f"下载论文失败 {arxiv_url}: {e}")
            
            return {
                "status": "error",
                "arxiv_url": arxiv_url,
                "message": error_msg,
                "error": str(e)
            }
    
    def download_papers_batch(self, arxiv_urls: List[str]) -> Dict[str, any]:
        """批量下载论文"""
        results = {
            "total": len(arxiv_urls),
            "success": 0,
            "exists": 0,
            "failed": 0,
            "details": []
        }
        
        print(f"📥 开始批量下载 {len(arxiv_urls)} 篇论文到 {self.papers_dir}/ 目录")
        print("=" * 60)
        
        for i, arxiv_url in enumerate(arxiv_urls, 1):
            print(f"\n📄 [{i}/{len(arxiv_urls)}] 处理: {arxiv_url}")
            
            result = self.download_single_paper(arxiv_url)
            results["details"].append(result)
            
            if result["status"] == "success":
                results["success"] += 1
                print(f"✅ {result['message']} ({result['size']} bytes)")
            elif result["status"] == "exists":
                results["exists"] += 1
                print(f"ℹ️  {result['message']} ({result['size']} bytes)")
            else:
                results["failed"] += 1
                print(f"❌ {result['message']}")
        
        # 显示总结
        print("\n" + "=" * 60)
        print("📊 下载结果总结:")
        print(f"  ✅ 新下载: {results['success']} 篇")
        print(f"  ℹ️  已存在: {results['exists']} 篇")
        print(f"  ❌ 失败: {results['failed']} 篇")
        print(f"  📝 总计: {results['total']} 篇")
        
        return results

def download_paper_list(paper_urls: List[str], papers_dir: str = "papers") -> Dict[str, any]:
    """
    简单的论文批量下载函数
    
    Args:
        paper_urls (List[str]): arXiv论文URL列表
        papers_dir (str): 下载目录，默认为papers
        
    Returns:
        Dict[str, any]: 下载结果统计
    """
    downloader = PaperDownloader(papers_dir)
    return downloader.download_papers_batch(paper_urls)

# 演示用法
if __name__ == "__main__":
    # 测试论文列表
    test_papers = [
        "https://arxiv.org/abs/1706.03762",  # Attention Is All You Need
        "https://arxiv.org/abs/1810.04805",  # BERT
        "https://arxiv.org/abs/2005.14165",  # GPT-3
    ]
    
    print("🚀 论文下载工具演示")
    result = download_paper_list(test_papers)
    
    print(f"\n🎯 下载完成！成功: {result['success']}, 已存在: {result['exists']}, 失败: {result['failed']}") 