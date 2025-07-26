from langchain_docling import DoclingLoader
import json
import logging
import re
import requests
import xml.etree.ElementTree as ET
from urllib.parse import urlparse
from pathlib import Path
import hashlib
from dataclasses import dataclass
from typing import Dict, List, Any, Optional
from datetime import datetime
import os

# 配置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')

@dataclass
class Paper:
    """论文数据类"""
    title: str
    authors: List[str]
    abstract: str
    journal: str
    year: int
    doi: str
    url: str
    pdf_url: str
    parsed_text: str = ""
    arxiv_id: str = ""
    citation_count: int = 0

class PDFTextExtractor:
    """PDF文本提取工具类 - 专门处理PDF文档的文本提取和元数据解析"""
    
    def __init__(self):
        self.logger = logging.getLogger(__name__)
        self.logger.info("PDF文本提取工具初始化成功")
    
    def download_pdf(self, pdf_url: str, arxiv_id: str = None) -> Optional[str]:
        """下载PDF文件到本地，使用arxiv_id命名"""
        try:
            # 创建下载目录
            download_dir = Path("downloads")
            download_dir.mkdir(exist_ok=True)
            
            # 生成文件名 - 优先使用arxiv_id，否则使用hash
            if arxiv_id:
                filename = f"{arxiv_id}.pdf"
            else:
                url_hash = hashlib.md5(pdf_url.encode()).hexdigest()[:8]
                filename = f"paper_{url_hash}.pdf"
                local_path = download_dir / filename
                
                # 如果文件已存在，直接返回路径
                if local_path.exists():
                    self.logger.info(f"PDF已存在: {local_path}")
                    return str(local_path)
                
                # 下载文件
                self.logger.info(f"下载PDF: {pdf_url}")
                headers = {
                    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
                }
                
                response = requests.get(pdf_url, headers=headers, timeout=30)
                response.raise_for_status()
                
                # 保存文件
                with open(local_path, 'wb') as f:
                    f.write(response.content)
                
                self.logger.info(f"PDF下载成功: {local_path}")
                return str(local_path)
            
        except Exception as e:
            self.logger.error(f"PDF下载失败: {e}")
            return None
    
    def load_paper_with_docling(self, file_path: str) -> List[Dict]:
        """使用Docling加载论文内容，如果已有解析结果则直接读取"""
        try:
            # 提取arxiv_id用于查找缓存
            arxiv_id = self.extract_arxiv_id(file_path)
            
            # 检查多个可能的缓存位置
            possible_cache_paths = []
            
            # 1. 同目录下的.md文件（优先）
            markdown_path = file_path + ".md"
            possible_cache_paths.append(markdown_path)
            
            # 2. downloads目录中的arXiv ID命名文件
            if arxiv_id:
                downloads_path = f"downloads/{arxiv_id}.pdf.md"
                possible_cache_paths.append(downloads_path)
            
            # 查找可用的缓存文件
            cached_content = None
            cache_source = None
            
            for cache_path in possible_cache_paths:
                if os.path.exists(cache_path):
                    self.logger.info(f"发现已有解析结果，直接读取: {cache_path}")
                    with open(cache_path, 'r', encoding='utf-8') as f:
                        cached_content = f.read()
                    cache_source = cache_path
                    break
            
            if cached_content:
                # 如果缓存不在当前文件旁边，复制一份到当前位置
                if cache_source != markdown_path:
                    with open(markdown_path, "w", encoding="utf-8") as f:
                        f.write(cached_content)
                    self.logger.info(f"缓存已复制到: {markdown_path}")
                
                # 返回与docling相同的格式
                paper_docs = [{
                    'content': cached_content,
                    'metadata': {
                        'source': file_path,
                        'cached': True,
                        'cache_source': cache_source
                    }
                }]
                self.logger.info(f"成功从缓存加载论文，长度: {len(cached_content)} 字符")
                return paper_docs
            
            # 如果没有缓存，使用Docling解析
            self.logger.info(f"使用Docling加载论文: {file_path}")
            loader = DoclingLoader(file_path=file_path)
            docs = loader.load()
            
            # 转换为字典格式
            paper_docs = []
            for doc in docs:
                paper_docs.append({
                    'content': doc.page_content,
                    'metadata': doc.metadata
                })
            
            # 合并所有内容并保存到markdown文件
            full_content = "\n".join([doc['content'] for doc in paper_docs])
            with open(markdown_path, "w", encoding="utf-8") as f:
                f.write(full_content)
            self.logger.info(f"解析结果已保存: {markdown_path}")
            
            # 如果有arxiv_id，也保存到downloads目录
            if arxiv_id:
                downloads_dir = Path("downloads")
                downloads_dir.mkdir(exist_ok=True)
                downloads_cache = downloads_dir / f"{arxiv_id}.pdf.md"
                if not downloads_cache.exists():
                    with open(downloads_cache, "w", encoding="utf-8") as f:
                        f.write(full_content)
                    self.logger.info(f"解析结果也已保存到: {downloads_cache}")
            
            self.logger.info(f"成功加载论文，共{len(paper_docs)}个文档片段")
            return paper_docs
            
        except Exception as e:
            self.logger.error(f"Docling加载论文失败: {e}")
            raise
    
    def _process_docs_in_chunks(self, docs: List) -> str:
        """分块处理文档，避免token长度超限"""
        all_text = []
        
        for doc in docs:
            content = doc.page_content if hasattr(doc, 'page_content') else str(doc)
            
            # 分块处理长文本（每块约4000字符）
            chunk_size = 4000
            if len(content) > chunk_size:
                chunks = [content[i:i+chunk_size] for i in range(0, len(content), chunk_size)]
                for i, chunk in enumerate(chunks):
                    all_text.append(f"## 文档片段 {i+1}\n\n{chunk}\n")
            else:
                all_text.append(content)
        
        return "\n".join(all_text)
    
    def extract_arxiv_id(self, file_path: str) -> str:
        """从文件路径中提取arxiv ID"""
        # 处理不同格式的arxiv链接和文件名
        patterns = [
            # 标准arXiv ID格式
            r'(?:arxiv[:\.]|arXiv[:\.])?(\d{4}\.\d{4,5}(?:v\d+)?)',
            # URL中的arXiv ID
            r'https?://arxiv\.org/(?:abs|pdf)/(\d{4}\.\d{4,5}(?:v\d+)?)',
            # 旧格式的arXiv ID (已不常用，但保留支持)
            r'(?:arxiv[:\.]|arXiv[:\.])([a-z-]+/\d{7}(?:v\d+)?)',
        ]
        
        for pattern in patterns:
            match = re.search(pattern, file_path, re.IGNORECASE)
            if match:
                arxiv_id = match.group(1)
                # 验证arXiv ID格式
                if re.match(r'\d{4}\.\d{4,5}(?:v\d+)?$', arxiv_id) or re.match(r'[a-z-]+/\d{7}(?:v\d+)?$', arxiv_id):
                    self.logger.info(f"提取到arXiv ID: {arxiv_id}")
                    return arxiv_id
        
        # 如果没有找到有效的arXiv ID，返回空字符串而不是hash
        self.logger.warning(f"未能从路径中提取有效的arXiv ID: {file_path}")
        return ""
    
    def _is_valid_arxiv_id(self, arxiv_id: str) -> bool:
        """验证arXiv ID格式是否有效"""
        if not arxiv_id:
            return False
        
        # 新格式: YYMM.NNNNN[vN] (如 1706.03762, 1706.03762v1)
        new_format = re.match(r'^\d{4}\.\d{4,5}(?:v\d+)?$', arxiv_id)
        
        # 旧格式: subject-class/YYMMnnn[vN] (如 cs/0001001, math/0309030v1)
        old_format = re.match(r'^[a-z-]+/\d{7}(?:v\d+)?$', arxiv_id)
        
        is_valid = bool(new_format or old_format)
        self.logger.debug(f"arXiv ID '{arxiv_id}' 格式验证: {'有效' if is_valid else '无效'}")
        
        return is_valid
    
    def extract_paper_metadata(self, docs: List[Dict], file_path: str) -> Dict[str, str]:
        """从文档中提取论文元数据"""
        # 合并所有文档内容
        full_content = "\n".join([doc['content'] for doc in docs])
        #print("内容测试",full_content)
        # 提取元数据
        metadata = {
            'arxiv_id': self.extract_arxiv_id(file_path),
            'title': self._extract_title(full_content),
            'authors': self._extract_authors(full_content),
            'abstract': self._extract_abstract(full_content),
            'year': self._extract_year(full_content),
            'pdf_url': file_path,
        }
        
        return metadata
    
    def _extract_title(self, content: str) -> str:
        """提取论文标题"""
        # 简单的标题提取逻辑
        lines = content.split('\n')
        for line in lines[:20]:  # 查看前20行
            line = line.strip()
            if len(line) > 10 and len(line) < 200 and not line.lower().startswith('abstract'):
                return line
        return "未知标题"
    
    def _extract_authors(self, content: str) -> List[str]:
        """提取作者信息，返回作者列表"""
        # 简单的作者提取逻辑
        lines = content.split('\n')
        for i, line in enumerate(lines[:30]):
            line = line.strip()
            if 'author' in line.lower() or '@' in line:
                # 尝试解析多个作者
                if ',' in line:
                    # 如果有逗号，按逗号分割作者
                    authors = [author.strip() for author in line.split(',') if author.strip()]
                    return authors
                elif 'and' in line.lower():
                    # 如果有"and"，按"and"分割作者
                    authors = [author.strip() for author in re.split(r'\s+and\s+', line, flags=re.IGNORECASE) if author.strip()]
                    return authors
                else:
                    # 单个作者或无法解析的情况
                    return [line.strip()]
        
        return ["未知作者"]
    
    def _extract_abstract(self, content: str) -> str:
        """提取摘要"""
        # 查找Abstract部分
        abstract_match = re.search(r'(?i)abstract\s*:?\s*(.*?)(?=\n\n|\nintroduction|\n1\.)', content, re.DOTALL)
        if abstract_match:
            return abstract_match.group(1).strip()[:1000]  # 限制长度
        return "未找到摘要"
    
    def _extract_year(self, content: str) -> int:
        """从PDF内容中提取发表年份"""
        # 多种年份提取模式
        year_patterns = [
            # 匹配常见的年份格式
            r'(?:published|submitted|accepted|presented|conference|proceedings).*?(\d{4})',
            r'(\d{4})\s*(?:conference|workshop|proceedings|journal)',
            r'(?:copyright|©).*?(\d{4})',
            r'(?:june|july|august|september|october|november|december|january|february|march|april|may)\s+(\d{4})',
            r'(?:jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\.?\s+(\d{4})',
            # 匹配arXiv格式的年份
            r'arXiv:(\d{4})\.\d+',
            # 通用4位数字年份（在合理范围内）
            r'\b(20[0-2][0-9]|19[789][0-9])\b'
        ]
        
        # 从内容的前1000个字符中查找年份
        search_content = content[:1000].lower()
        
        for pattern in year_patterns:
            matches = re.findall(pattern, search_content, re.IGNORECASE)
            if matches:
                for match in matches:
                    year = int(match)
                    # 验证年份的合理性（1990-2030）
                    if 1990 <= year <= 2030:
                        return year
        
        # 如果都没找到，尝试从文件名中提取
        filename = Path(self.extract_arxiv_id(content)).name if content else ""
        filename_match = re.search(r'(20[0-2][0-9]|19[789][0-9])', filename)
        if filename_match:
            return int(filename_match.group(1))
        
        # 最后默认返回当前年份
        return datetime.now().year
    
    def fetch_arxiv_metadata(self, arxiv_id: str) -> Optional[Paper]:
        """从arXiv API获取论文元数据"""
        try:
            # arXiv API URL - 使用http协议
            api_url = f"http://export.arxiv.org/api/query?id_list={arxiv_id}"
            
            self.logger.info(f"获取arXiv元数据: {arxiv_id}")
            self.logger.info(f"API URL: {api_url}")
            
            # 设置较短的超时时间和重试机制
            headers = {
                'User-Agent': 'Mozilla/5.0 (Linux; x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
            }
            
            response = requests.get(api_url, headers=headers, timeout=15)
            response.raise_for_status()
            
            # 检查响应内容
            if not response.content:
                self.logger.warning(f"arXiv API返回空内容: {arxiv_id}")
                return None
            
            # 解析XML响应
            root = ET.fromstring(response.content)
            namespaces = {
                'atom': 'http://www.w3.org/2005/Atom',
                'arxiv': 'http://arxiv.org/schemas/atom'
            }
            
            # 找到第一个entry
            entry = root.find('atom:entry', namespaces)
            if entry is None:
                self.logger.warning(f"未找到arXiv条目: {arxiv_id}")
                return None
                
            # 解析条目
            return self._parse_entry_element(entry, namespaces)
            
        except requests.exceptions.Timeout:
            self.logger.error(f"arXiv API请求超时: {arxiv_id}")
            return None
        except requests.exceptions.ConnectionError:
            self.logger.error(f"无法连接arXiv API: {arxiv_id}")
            return None
        except requests.exceptions.HTTPError as e:
            self.logger.error(f"arXiv API HTTP错误: {e}")
            return None
        except ET.ParseError as e:
            self.logger.error(f"arXiv API响应解析失败: {e}")
            return None
        except Exception as e:
            self.logger.error(f"获取arXiv元数据失败: {e}")
            return None
    
    def _parse_entry_element(self, entry, namespaces) -> Optional[Paper]:
        """解析XML entry元素"""
        try:
            # 获取标题
            title_elem = entry.find('atom:title', namespaces)
            if title_elem is None:
                return None
            title = title_elem.text.strip().replace('\n', ' ')
            
            # 获取摘要
            summary_elem = entry.find('atom:summary', namespaces)
            abstract = summary_elem.text.strip().replace('\n', ' ') if summary_elem is not None else ""
            
            # 获取作者
            authors = []
            for author_elem in entry.findall('atom:author', namespaces):
                name_elem = author_elem.find('atom:name', namespaces)
                if name_elem is not None:
                    authors.append(name_elem.text.strip())
            
            # 获取发布日期
            published_elem = entry.find('atom:published', namespaces)
            year = 1970  # 默认年份
            if published_elem is not None:
                try:
                    year = int(published_elem.text[:4])
                except (ValueError, IndexError):
                    pass
            
            # 获取DOI
            doi_elem = entry.find('arxiv:doi', namespaces)
            doi = doi_elem.text.strip() if doi_elem is not None else ""
            
            # 获取ArXiv ID和链接  
            id_elem = entry.find('atom:id', namespaces)
            url = id_elem.text.strip() if id_elem is not None else ""
            arxiv_id = self.extract_arxiv_id(url)
            
            # 构建PDF链接
            pdf_url = ""
            for link in entry.findall('atom:link', namespaces):
                if link.get('type') == 'application/pdf':
                    pdf_url = link.get('href', '')
                    break
            
            # 如果没有找到PDF链接，尝试构建标准arXiv PDF链接
            if not pdf_url and arxiv_id:
                pdf_url = f"https://arxiv.org/pdf/{arxiv_id}.pdf"
            
            # 下载PDF并解析
            parsed_text = ""
            if pdf_url:
                local_pdf = self.download_pdf(pdf_url, arxiv_id)
                if local_pdf:
                    try:
                        # 检查多个可能的缓存位置
                        possible_cache_paths = []
                        
                        # 1. 同目录下的.md文件
                        markdown_path = local_pdf + ".md"
                        possible_cache_paths.append(markdown_path)
                        
                        # 2. downloads目录中的arXiv ID命名文件
                        if arxiv_id:
                            downloads_cache = f"downloads/{arxiv_id}.pdf.md"
                            possible_cache_paths.append(downloads_cache)
                        
                        # 查找可用的缓存文件
                        cached_content = None
                        cache_source = None
                        
                        for cache_path in possible_cache_paths:
                            if os.path.exists(cache_path):
                                self.logger.info(f"发现已有解析结果，直接读取: {cache_path}")
                                with open(cache_path, 'r', encoding='utf-8') as f:
                                    cached_content = f.read()
                                cache_source = cache_path
                                break
                        
                        if cached_content:
                            parsed_text = cached_content
                            # 如果缓存不在当前文件旁边，复制一份
                            if cache_source != markdown_path:
                                with open(markdown_path, "w", encoding="utf-8") as f:
                                    f.write(parsed_text)
                                self.logger.info(f"缓存已复制到: {markdown_path}")
                        else:
                            # 如果没有缓存，进行解析
                            loader = DoclingLoader(file_path=local_pdf)
                            docs = loader.load()
                            
                            # 分段处理，避免token长度超限
                            parsed_text = self._process_docs_in_chunks(docs)
                            
                            # 存储解析结果到本地
                            with open(markdown_path, "w", encoding="utf-8") as f:
                                f.write(parsed_text)
                            self.logger.info(f"解析结果已保存: {markdown_path}")
                                
                            # 如果有arxiv_id，也保存到downloads目录
                            if arxiv_id:
                                downloads_dir = Path("downloads")
                                downloads_dir.mkdir(exist_ok=True)
                                downloads_cache = downloads_dir / f"{arxiv_id}.pdf.md"
                                if not downloads_cache.exists():
                                    with open(downloads_cache, "w", encoding="utf-8") as f:
                                        f.write(parsed_text)
                                    self.logger.info(f"解析结果也已保存到: {downloads_cache}")
                        
                    except Exception as e:
                        self.logger.error(f"PDF解析失败: {e}")
            
            # 创建Paper对象
            paper = Paper(
                title=title,
                authors=authors,
                abstract=abstract,
                journal="arXiv preprint",
                year=year,
                doi=doi,
                url=url,
                pdf_url=pdf_url,
                parsed_text=parsed_text,
                arxiv_id=arxiv_id
            )
            
            return paper
            
        except Exception as e:
            self.logger.error(f"解析arXiv条目失败: {e}")
            return None
    
    def extract_text_from_pdf(self, file_path: str) -> Dict[str, Any]:
        """从PDF文件中提取文本和元数据的主要方法"""
        try:
            self.logger.info(f"开始提取PDF文本: {file_path}")
            
            # 使用Docling加载文档
            docs = self.load_paper_with_docling(file_path)
            
            # 合并文档内容
            full_content = "\n".join([doc['content'] for doc in docs])
            
            # 首先尝试从文件路径提取arXiv ID
            arxiv_id = self.extract_arxiv_id(file_path)
            
            # 尝试从arXiv API获取元数据
            paper_from_arxiv = None
            if arxiv_id and self._is_valid_arxiv_id(arxiv_id):
                self.logger.info(f"尝试从arXiv API获取元数据: {arxiv_id}")
                paper_from_arxiv = self.fetch_arxiv_metadata(arxiv_id)
            elif arxiv_id:
                self.logger.warning(f"无效的arXiv ID格式，跳过API调用: {arxiv_id}")
            else:
                self.logger.info("未找到arXiv ID，直接使用PDF内容提取")
            
            # 创建Paper对象
            if paper_from_arxiv:
                # 使用arXiv API获取的准确元数据，但使用本地解析的文本内容
                self.logger.info("✅ 使用arXiv API元数据")
                paper = Paper(
                    title=paper_from_arxiv.title,
                    authors=paper_from_arxiv.authors,
                    abstract=paper_from_arxiv.abstract,
                    journal=paper_from_arxiv.journal,
                    year=paper_from_arxiv.year,
                    doi=paper_from_arxiv.doi,
                    url=paper_from_arxiv.url,
                    pdf_url=file_path,  # 使用本地文件路径
                    parsed_text=full_content,  # 使用本地解析的文本
                    arxiv_id=paper_from_arxiv.arxiv_id
                )
            else:
                # 回退到从PDF内容提取元数据
                self.logger.info("⚠️ 无法从arXiv获取元数据，使用PDF内容提取")
                metadata = self.extract_paper_metadata(docs, file_path)
                #MVP阶段使用固定引用量字典
                cite_dict = {
                    "1706.03762": 186267,
                    "1810.04805": 139994,
                    "2005.14165": 50786,
                    "2201.11903": 17271,
                    "2210.03629": 3815,
                    "2302.04761": 2027,
                    "2304.03442": 2604,
                    "2310.0856": 193,
                    "2304.11477": 519,
                    "2303.03378": 2191,
                }
                paper = Paper(
                    title=metadata['title'],
                    authors=metadata['authors'],
                    abstract=metadata['abstract'],
                    journal="Unknown",
                    year=metadata['year'],
                    doi="",
                    url=file_path,
                    pdf_url=metadata['pdf_url'],
                    parsed_text=full_content,
                    arxiv_id=metadata['arxiv_id'],
                    citation_count=cite_dict[metadata['arxiv_id']]
                )
            
            result = {
                "status": "success",
                "paper": {
                    "arxiv_id": paper.arxiv_id,
                    "title": paper.title,
                    "authors": paper.authors,
                    "abstract": paper.abstract,
                    "journal": paper.journal,
                    "year": paper.year,
                    "doi": paper.doi,
                    "url": paper.url,
                    "pdf_url": paper.pdf_url,
                    "parsed_text": paper.parsed_text,
                    "parsed_text_length": len(paper.parsed_text),
                    "citation_count": paper.citation_count
                },
                "message": "PDF文本提取完成",
                "metadata_source": "arxiv_api" if paper_from_arxiv else "pdf_content"
            }
            
            self.logger.info(f"PDF文本提取成功: {paper.arxiv_id}")
            return result
            
        except Exception as e:
            self.logger.error(f"PDF文本提取失败: {e}")
            return {
                "status": "error",
                "message": f"PDF文本提取失败: {str(e)}"
            }

# 演示用法
if __name__ == "__main__":
    # 创建PDF文本提取器
    extractor = PDFTextExtractor()

    paper_list = [
        "1706.03762v7.pdf",

        ]
    # 遍历papers目录下的所有PDF文件
    papers_dir = "/home/bugsmith/paperplay/papers"
    for file in os.listdir(papers_dir):
        if file.endswith(".pdf"):
            paper_path = os.path.join(papers_dir, file)
            paper_list.append(paper_path)
    
    # 提取示例PDF
    FILE_PATH = "/home/bugsmith/paperplay/papers/1706.03762v7.pdf"
    result = extractor.extract_text_from_pdf(FILE_PATH)
    
    print("提取结果:")
    print(json.dumps(result, ensure_ascii=False, indent=2)) 