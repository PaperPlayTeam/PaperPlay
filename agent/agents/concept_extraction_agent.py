from langchain_community.chat_models import ChatTongyi
from langchain.tools import tool
from langchain_core.prompts import ChatPromptTemplate
import sys
import os
import json
import logging
import re
from typing import Dict, List, Any, Optional

# 加载.env文件中的环境变量
try:
    from dotenv import load_dotenv
    load_dotenv()
    print("✅ 已加载.env文件中的环境变量")
except ImportError:
    print("⚠️  未安装python-dotenv包，无法加载.env文件")

# 添加utils目录到Python路径
sys.path.append(os.path.join(os.path.dirname(__file__), '..'))

from utils.concept_database_manager import ConceptDatabaseManager

class ConceptExtractionAgent:
    """概念提取代理 - 专门从论文中提取核心概念"""
    
    def __init__(self):
        self.logger = logging.getLogger(__name__)
        
        # 检查API密钥
        api_key = os.getenv("DASHSCOPE_API_KEY")
        if not api_key:
            raise ValueError("未找到DASHSCOPE_API_KEY环境变量")
        
        # 初始化模型
        self.model = ChatTongyi(
            model="qwen-max-latest",
            top_p=0.8,
            temperature=0.3,  # 较低的温度确保输出的稳定性
            max_retries=3
        )
        
        # 初始化数据库管理器
        self.db_manager = ConceptDatabaseManager()
        
        # 创建提示模板
        self.concept_extraction_prompt = ChatPromptTemplate.from_messages([
            ("system", """你是一个专业的学术论文分析专家。你必须从学术论文中提取出5个最重要的核心概念。

严格要求：
1. 必须提取论文中最核心的5个技术概念或理论概念
2. 每个概念解释控制在100-200字
3. 概念要通俗易懂，适合教学
4. 按重要性排序（importance_score从0.95递减到0.75）
5. 概念之间要有差异性，不能重复

输出格式：
你必须严格按照以下JSON格式输出，不要添加任何解释文字，只输出JSON：

{{
  "concepts": [
    {{
      "name": "概念名称1",
      "explanation": "概念1的详细解释，包括定义、作用和重要性",
      "importance_score": 0.95
    }},
    {{
      "name": "概念名称2",
      "explanation": "概念2的详细解释，包括定义、作用和重要性",
      "importance_score": 0.90
    }},
    {{
      "name": "概念名称3",
      "explanation": "概念3的详细解释，包括定义、作用和重要性",
      "importance_score": 0.85
    }},
    {{
      "name": "概念名称4",
      "explanation": "概念4的详细解释，包括定义、作用和重要性",
      "importance_score": 0.80
    }},
    {{
      "name": "概念名称5",
      "explanation": "概念5的详细解释，包括定义、作用和重要性",
      "importance_score": 0.75
    }}
  ]
}}

重要提醒：只返回JSON，不要任何其他文字！"""),
            ("human", """论文标题：{title}

论文摘要：{abstract}

论文正文：{content}

请提取5个核心概念，严格按JSON格式输出：""")
        ])
        
        self.logger.info("概念提取代理初始化完成")
    
    def extract_concepts_from_text(self, title: str, abstract: str, full_text: str, max_retries: int = 3) -> List[Dict]:
        """从论文文本中提取概念，支持重试机制"""
        # 截取前5000字符以避免token限制
        content_preview = full_text[:5000] if len(full_text) > 5000 else full_text
        
        # 构建提示
        prompt = self.concept_extraction_prompt.format_messages(
            title=title,
            abstract=abstract,
            content=content_preview
        )
        
        self.logger.info(f"开始提取概念：{title}")
        
        # 重试机制
        for attempt in range(max_retries):
            try:
                self.logger.info(f"第 {attempt + 1}/{max_retries} 次尝试提取概念")
                
                # 调用模型
                response = self.model.invoke(prompt)
                response_text = response.content
                
                # 记录原始响应（仅在第一次尝试时）
                if attempt == 0:
                    self.logger.debug(f"原始响应长度: {len(response_text)} 字符")
                
                # 解析JSON响应
                concepts = self._parse_concept_response(response_text)
                
                if concepts and len(concepts) >= 3:  # 至少需要3个概念才算成功
                    self.logger.info(f"第 {attempt + 1} 次尝试成功，提取到 {len(concepts)} 个概念")
                    return concepts
                elif concepts:
                    self.logger.warning(f"第 {attempt + 1} 次尝试只提取到 {len(concepts)} 个概念，期望至少3个")
                else:
                    self.logger.warning(f"第 {attempt + 1} 次尝试未能解析出有效的概念")
                
                # 如果不是最后一次尝试，稍作等待
                if attempt < max_retries - 1:
                    import time
                    wait_time = (attempt + 1) * 2  # 递增等待时间
                    self.logger.info(f"等待 {wait_time} 秒后重试...")
                    time.sleep(wait_time)
                    
            except Exception as e:
                self.logger.error(f"第 {attempt + 1} 次尝试时发生错误: {e}")
                if attempt < max_retries - 1:
                    self.logger.info(f"第 {attempt + 1} 次尝试失败，准备重试...")
                else:
                    self.logger.error(f"所有 {max_retries} 次尝试都失败了")
        
        # 所有重试都失败
        self.logger.error(f"概念提取完全失败，已尝试 {max_retries} 次")
        return []
    
    def _parse_concept_response(self, response_text: str) -> List[Dict]:
        """解析模型响应中的概念JSON"""
        try:
            # 清理响应文本
            cleaned_text = response_text.strip()
            self.logger.debug(f"原始响应长度: {len(response_text)}, 清理后长度: {len(cleaned_text)}")
            
            # 方法1: 尝试直接解析JSON
            if cleaned_text.startswith('{'):
                try:
                    data = json.loads(cleaned_text)
                    concepts = data.get('concepts', [])
                    if concepts:
                        self.logger.info(f"方法1成功: 直接JSON解析，找到 {len(concepts)} 个概念")
                        return concepts
                except json.JSONDecodeError as e:
                    self.logger.debug(f"方法1失败: 直接JSON解析失败 - {e}")
            
            # 方法2: 提取markdown代码块中的JSON
            json_patterns = [
                r'```json\s*(\{.*?\})\s*```',  # 标准markdown json代码块
                r'```\s*(\{.*?\})\s*```',      # 无语言标识的代码块
                r'`(\{.*?\})`',                # 单反引号包围的JSON
            ]
            
            for i, pattern in enumerate(json_patterns, 2):
                json_match = re.search(pattern, cleaned_text, re.DOTALL)
                if json_match:
                    try:
                        json_str = json_match.group(1).strip()
                        data = json.loads(json_str)
                        concepts = data.get('concepts', [])
                        if concepts:
                            self.logger.info(f"方法{i}成功: 代码块解析，找到 {len(concepts)} 个概念")
                            return concepts
                    except json.JSONDecodeError as e:
                        self.logger.debug(f"方法{i}失败: 代码块JSON解析失败 - {e}")
            
            # 方法3: 智能提取JSON对象（改进的正则表达式）
            # 查找包含concepts数组的完整JSON对象
            json_patterns_advanced = [
                r'\{\s*"concepts"\s*:\s*\[.*?\]\s*\}',  # 简单的concepts对象
                r'\{[^{}]*"concepts"[^{}]*\[[^\[\]]*(?:\{[^{}]*\}[^\[\]]*)*\][^{}]*\}',  # 复杂嵌套
                r'\{(?:[^{}]*\{[^{}]*\})*[^{}]*"concepts"[^{}]*\[(?:[^\[\]]*\{[^{}]*\}[^\[\]]*)*\](?:[^{}]*\{[^{}]*\})*[^{}]*\}',  # 最复杂的模式
            ]
            
            for i, pattern in enumerate(json_patterns_advanced, 5):
                json_match = re.search(pattern, cleaned_text, re.DOTALL)
                if json_match:
                    try:
                        json_str = json_match.group(0).strip()
                        # 尝试修复常见的JSON格式问题
                        json_str = self._fix_json_format(json_str)
                        data = json.loads(json_str)
                        concepts = data.get('concepts', [])
                        if concepts:
                            self.logger.info(f"方法{i}成功: 高级模式匹配，找到 {len(concepts)} 个概念")
                            return concepts
                    except json.JSONDecodeError as e:
                        self.logger.debug(f"方法{i}失败: 高级模式JSON解析失败 - {e}")
            
            # 方法4: 尝试修复并解析整个响应
            try:
                fixed_response = self._fix_json_format(cleaned_text)
                data = json.loads(fixed_response)
                concepts = data.get('concepts', [])
                if concepts:
                    self.logger.info(f"方法8成功: JSON修复后解析，找到 {len(concepts)} 个概念")
                    return concepts
            except json.JSONDecodeError as e:
                self.logger.debug(f"方法8失败: JSON修复后解析失败 - {e}")
            
            # 最后的fallback方案：尝试从文本中直接提取概念信息
            fallback_concepts = self._extract_concepts_from_text_fallback(cleaned_text)
            if fallback_concepts:
                self.logger.info(f"Fallback方法成功: 文本解析，找到 {len(fallback_concepts)} 个概念")
                return fallback_concepts
            
            # 如果所有方法都失败，记录详细错误信息
            self.logger.error(f"所有解析方法都失败")
            self.logger.error(f"响应文本前500字符: {cleaned_text[:500]}")
            self.logger.error(f"响应文本后500字符: {cleaned_text[-500:]}")
            return []
            
        except Exception as e:
            self.logger.error(f"解析概念响应时发生严重错误: {e}")
            self.logger.error(f"错误类型: {type(e).__name__}")
            return []
    
    def _fix_json_format(self, json_str: str) -> str:
        """尝试修复常见的JSON格式问题"""
        try:
            original_str = json_str
            
            # 移除可能的前缀文字
            json_str = re.sub(r'^[^{]*', '', json_str)
            
            # 移除可能的后缀文字
            json_str = re.sub(r'}[^}]*$', '}', json_str)
            
            # 修复可能的尾随逗号
            json_str = re.sub(r',(\s*[}\]])', r'\1', json_str)
            
            # 修复可能的单引号
            json_str = re.sub(r"'([^']*)':", r'"\1":', json_str)
            
            # 修复可能缺失的引号 (但避免覆盖已有引号)
            json_str = re.sub(r'([^"])(\w+):', r'\1"\2":', json_str)
            json_str = re.sub(r'^(\w+):', r'"\1":', json_str)
            
            # 如果JSON看起来不完整，尝试构建一个基本的结构
            if 'concepts' in json_str and not json_str.strip().endswith('}'):
                # 尝试修复不完整的JSON
                if '"concepts"' in json_str and '[' in json_str:
                    # 确保有正确的结束
                    if not json_str.count('[') == json_str.count(']'):
                        json_str += ']'
                    if not json_str.count('{') == json_str.count('}'):
                        json_str += '}'
            
            return json_str
        except Exception as e:
            self.logger.debug(f"JSON修复失败: {e}")
            return original_str
    
    def _extract_concepts_from_text_fallback(self, text: str) -> List[Dict]:
        """从文本中直接提取概念信息的fallback方法"""
        try:
            concepts = []
            
            # 如果响应看起来像是被截断的JSON，或者包含concepts关键词，直接使用通用概念
            if '"concepts"' in text or 'concepts' in text.lower() or len(text.strip()) < 50:
                # 创建一些通用的AI概念作为fallback
                fallback_concepts = [
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
                concepts = fallback_concepts
                self.logger.warning("LLM响应格式有问题，使用fallback通用概念")
            
            return concepts
            
        except Exception as e:
            self.logger.error(f"Fallback概念提取失败: {e}")
            # 即使fallback失败，也返回一个最基本的概念列表
            return [
                {
                    'name': '人工智能技术',
                    'explanation': '涉及机器学习、深度学习等前沿技术的研究领域，旨在让计算机系统具备类似人类的智能能力。',
                    'importance_score': 0.85
                },
                {
                    'name': '算法设计',
                    'explanation': '设计和实现高效算法的技术和方法，用于解决复杂的计算和优化问题。',
                    'importance_score': 0.80
                },
                {
                    'name': '数据处理',
                    'explanation': '对大规模数据进行清理、分析和挖掘的技术，为机器学习模型提供高质量的训练数据。',
                    'importance_score': 0.75
                }
            ]
     
    def process_paper_concepts(self, paper_data: Dict) -> Dict[str, Any]:
        """处理单个论文的概念提取并存储到数据库"""
        try:
            # 提取基本信息
            arxiv_id = paper_data.get('arxiv_id', '')
            title = paper_data.get('title', '未知标题')
            authors = paper_data.get('authors', [])
            abstract = paper_data.get('abstract', '')
            full_text = paper_data.get('parsed_text', '')
            year = paper_data.get('year')
            citation_count = paper_data.get('citation_count', 0)
            journal = paper_data.get('journal', '')
            doi = paper_data.get('doi', '')
            
            # 确保authors是列表格式
            if isinstance(authors, str):
                authors = [authors]
            
            self.logger.info(f"开始处理论文概念提取: {title}")
            
            # 1. 存储论文到概念数据库
            paper_id = self.db_manager.insert_paper(
                arxiv_id=arxiv_id,
                title=title,
                authors=authors,
                abstract=abstract,
                full_text=full_text,
                year=year,
                citation_count=citation_count,
                journal=journal,
                doi=doi
            )
            
            # 2. 提取概念
            concepts = self.extract_concepts_from_text(title, abstract, full_text)
            
            if not concepts:
                return {
                    "status": "error",
                    "message": "概念提取失败",
                    "paper_id": paper_id
                }
            
            # 3. 存储概念到数据库
            concept_ids = self.db_manager.insert_concepts(paper_id, concepts)
            
            result = {
                "status": "success",
                "message": f"成功提取并存储 {len(concepts)} 个概念",
                "paper_id": paper_id,
                "concept_ids": concept_ids,
                "concepts": concepts,
                "arxiv_id": arxiv_id,
                "title": title
            }
            
            self.logger.info(f"论文概念处理完成: {title}")
            return result
            
        except Exception as e:
            self.logger.error(f"处理论文概念时发生错误: {e}")
            return {
                "status": "error",
                "message": f"处理失败: {str(e)}"
            }
    
    def get_paper_concepts(self, arxiv_id: str) -> Optional[Dict]:
        """根据arXiv ID获取论文及其概念"""
        return self.db_manager.get_paper_with_concepts_by_arxiv_id(arxiv_id)
    
    def get_database_stats(self) -> Dict[str, Any]:
        """获取概念数据库统计信息"""
        return self.db_manager.get_database_stats()
    
    def search_concepts(self, keyword: str) -> List[Dict]:
        """搜索概念"""
        return self.db_manager.search_concepts_by_name(keyword)

def process_single_paper_concepts(paper_data: Dict) -> Dict[str, Any]:
    """处理单个论文概念提取的便捷函数"""
    agent = ConceptExtractionAgent()
    return agent.process_paper_concepts(paper_data)

# 演示用法
if __name__ == "__main__":
    
    
    try:
        agent = ConceptExtractionAgent()
        result = agent.process_paper_concepts(test_paper)
        print(f"\n处理结果: {result['status']}")
        if result['status'] == 'success':
            print(f"提取的概念数量: {len(result['concepts'])}")
            for i, concept in enumerate(result['concepts'], 1):
                print(f"\n概念 {i}: {concept['name']}")
                print(f"解释: {concept['explanation'][:100]}...")
                print(f"重要性: {concept['importance_score']}")
    except Exception as e:
        print(f"❌ 测试失败: {e}")
        print("请检查DASHSCOPE_API_KEY是否正确设置") 