#!/usr/bin/env python3
"""
从concepts.json文件生成类比引入式问题并存储到数据库
为每个概念生成引入题+概念题的配对问题
"""

import os
import sys
import json
import sqlite3
import logging
import uuid
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from datetime import datetime

# 添加项目根目录到Python路径
project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, project_root)

from langchain_community.chat_models import ChatTongyi
from langchain_core.prompts import ChatPromptTemplate

# 加载环境变量
try:
    from dotenv import load_dotenv
    load_dotenv()
    print("✅ 已加载.env文件中的环境变量")
except ImportError:
    print("⚠️  未安装python-dotenv包，无法加载.env文件")

# 设置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

class AnalogicalQuestionGenerator:
    """类比引入式问题生成器"""
    
    def __init__(self):
        # 检查API密钥
        api_key = os.getenv("DASHSCOPE_API_KEY")
        if not api_key:
            raise ValueError("未找到DASHSCOPE_API_KEY环境变量")
        
        # 初始化LLM模型
        self.model = ChatTongyi(
            model="qwen-max-latest",
            top_p=0.8,
            temperature=0.7,  
            max_retries=3
        )
        
        # 数据库路径
        self.db_path = os.path.join(project_root, 'sqlite', 'paperplay.db')
        
        # 创建问题生成的prompt模板
        self.question_prompt = ChatPromptTemplate.from_messages([
            ("system", """你是一个优秀的教育专家，专门设计类比引入式题目来帮助学习者理解复杂的计算机概念。

你的任务是基于给定的概念信息，生成一道完整的类比引入式题目，包含两个部分：

1. **引入题（生活类比）**：用生活中常见的场景类比计算机概念，提出问题并给出4个选项
2. **概念题（技术解释）**：解释计算机概念后，提出直接相关的问题并给出4个选项

**重要要求**：
- 引入题应构建**普适性的日常情境**，避免使用如“某某人”等具体代词或名字，以确保所有学习者都能产生共鸣。
- 引入题和概念题的正确答案必须在同一个选项位置（如都是B）
- 两个问题的逻辑必须高度对应
- 类比要贴切、自然、易懂
- 选项要有合理的干扰性
- 语言要生动、准确

请严格按照以下JSON格式输出：

{{
  "lead_in_question": "生活化的引入问题描述",
  "lead_in_options": [
    "A. 选项A内容",
    "B. 选项B内容", 
    "C. 选项C内容",
    "D. 选项D内容"
  ],
  "concept_explanation": "概念解释段落，连接类比和技术概念",
  "lead_in_question": "计算机概念相关的问题",
  "concept_options": [
    "A. 选项A内容",
    "B. 选项B内容",
    "C. 选项C内容", 
    "D. 选项D内容"
  ],
  "correct_option": "B",
  "explanation": "为什么这个选项是正确的，解释类比和概念的对应关系"
}}

只返回JSON，不要其他内容！"""),
            ("human", """论文标题：{paper_title}
概念名称：{concept_name}
概念解释：{concept_explanation}
重要性评分：{importance_score}

请基于这个概念生成一道类比引入式题目：""")
        ])
        
        logger.info("类比问题生成器初始化完成")

    def generate_question_for_concept(self, paper_title: str, concept: Dict, max_retries: int = 3) -> Optional[Dict]:
        """为单个概念生成类比引入式问题"""
        concept_name = concept.get('name', '')
        concept_explanation = concept.get('explanation', '')
        importance_score = concept.get('importance_score', 0)
        
        logger.info(f"开始为概念生成问题: {concept_name}")
        
        # 重试机制
        for attempt in range(max_retries):
            try:
                logger.info(f"第 {attempt + 1}/{max_retries} 次尝试生成问题")
                
                # 构建prompt
                prompt = self.question_prompt.format_messages(
                    paper_title=paper_title,
                    concept_name=concept_name,
                    concept_explanation=concept_explanation,
                    importance_score=importance_score
                )
                
                # 调用LLM
                response = self.model.invoke(prompt)
                response_text = response.content.strip()
                
                # 解析JSON响应
                question_data = self._parse_question_response(response_text)
                
                if question_data and self._validate_question_structure(question_data):
                    logger.info(f"第 {attempt + 1} 次尝试成功生成问题")
                    return question_data
                else:
                    logger.warning(f"第 {attempt + 1} 次尝试生成的问题格式不正确")
                
                # 如果不是最后一次尝试，稍作等待
                if attempt < max_retries - 1:
                    import time
                    wait_time = (attempt + 1) * 2
                    logger.info(f"等待 {wait_time} 秒后重试...")
                    time.sleep(wait_time)
                    
            except Exception as e:
                logger.error(f"第 {attempt + 1} 次尝试时发生错误: {e}")
                if attempt < max_retries - 1:
                    logger.info(f"第 {attempt + 1} 次尝试失败，准备重试...")
                else:
                    logger.error(f"所有 {max_retries} 次尝试都失败了")
        
        # 所有重试都失败，返回fallback问题
        logger.warning(f"为概念 {concept_name} 生成问题失败，使用fallback")
        return self._create_fallback_question(concept_name, concept_explanation)

    def _parse_question_response(self, response_text: str) -> Optional[Dict]:
        """解析LLM响应中的问题JSON"""
        try:
            # 清理响应文本
            cleaned_text = response_text.strip()
            
            # 尝试直接解析JSON
            if cleaned_text.startswith('{'):
                try:
                    data = json.loads(cleaned_text)
                    return data
                except json.JSONDecodeError:
                    pass
            
            # 尝试提取markdown代码块中的JSON
            import re
            json_patterns = [
                r'```json\s*(\{.*?\})\s*```',
                r'```\s*(\{.*?\})\s*```',
                r'`(\{.*?\})`',
            ]
            
            for pattern in json_patterns:
                json_match = re.search(pattern, cleaned_text, re.DOTALL)
                if json_match:
                    try:
                        json_str = json_match.group(1).strip()
                        data = json.loads(json_str)
                        return data
                    except json.JSONDecodeError:
                        continue
            
            logger.error(f"无法解析LLM响应为JSON: {cleaned_text[:200]}...")
            return None
            
        except Exception as e:
            logger.error(f"解析问题响应时发生错误: {e}")
            return None

    def _validate_question_structure(self, question_data: Dict) -> bool:
        """验证问题数据结构的完整性"""
        required_fields = [
            'lead_in_question', 'lead_in_options', 
            'concept_explanation', 'concept_question', 'concept_options',
            'correct_option', 'explanation'
        ]
        
        for field in required_fields:
            if field not in question_data:
                logger.warning(f"问题数据缺少必需字段: {field}")
                return False
        
        # 验证选项数量
        if len(question_data.get('lead_in_options', [])) != 4:
            logger.warning("引入题选项数量不是4个")
            return False
            
        if len(question_data.get('concept_options', [])) != 4:
            logger.warning("概念题选项数量不是4个")
            return False
        
        # 验证正确选项格式
        correct_option = question_data.get('correct_option', '')
        if correct_option not in ['A', 'B', 'C', 'D']:
            logger.warning(f"正确选项格式不正确: {correct_option}")
            return False
        
        return True

    def _create_fallback_question(self, concept_name: str, concept_explanation: str) -> Dict:
        """创建fallback问题（当LLM生成失败时使用）"""
        return {
            "lead_in_question": f"想象你需要向朋友解释一个复杂的概念：'{concept_name}'，你会采用什么方法？",
            "lead_in_options": [
                "A. 直接背诵专业定义，让朋友自己理解",
                "B. 用生活中的例子来类比，让复杂概念变得容易理解",
                "C. 要求朋友先学会所有相关的基础知识",
                "D. 画一个复杂的技术图表来说明"
            ],
            "concept_explanation": f"在计算机科学中，{concept_name}是一个重要概念。{concept_explanation[:100]}...",
            "concept_question": f"关于{concept_name}的核心作用，以下哪项描述最准确？",
            "concept_options": [
                "A. 主要用于提高系统的运行速度",
                "B. 通过特定的机制和方法，帮助系统更好地理解和处理复杂信息",
                "C. 仅用于数据存储和管理",
                "D. 专门处理用户界面的显示问题"
            ],
            "correct_option": "B",
            "explanation": f"就像用生活例子来解释复杂概念一样，{concept_name}通过其特定的机制帮助计算机系统更好地理解和处理信息。"
        }

class DatabaseManager:
    """数据库管理器"""
    
    def __init__(self, db_path: str):
        self.db_path = db_path
        logger.info(f"数据库管理器初始化: {db_path}")

    def get_connection(self):
        """获取数据库连接"""
        return sqlite3.connect(self.db_path)

    def ensure_subject_exists(self) -> str:
        """确保机器学习学科存在，返回subject_id"""
        subject_id = "ml_ai_subject"
        subject_name = "机器学习与人工智能"
        current_time = int(datetime.now().timestamp())
        
        with self.get_connection() as conn:
            cursor = conn.cursor()
            
            # 检查是否已存在
            cursor.execute("SELECT id FROM subjects WHERE id = ?", (subject_id,))
            if cursor.fetchone():
                logger.info(f"学科已存在: {subject_name}")
                return subject_id
            
            # 创建学科
            cursor.execute("""
                INSERT INTO subjects (id, name, description, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?)
            """, (subject_id, subject_name, "机器学习、深度学习等AI相关论文", current_time, current_time))
            
            conn.commit()
            logger.info(f"创建新学科: {subject_name}")
            return subject_id

    def insert_paper_and_level(self, subject_id: str, paper_info: Dict) -> Tuple[str, str]:
        """插入论文和对应的关卡，返回(paper_id, level_id)"""
        paper_id = f"paper_{paper_info['arxiv_id']}"
        level_id = f"level_{paper_info['arxiv_id']}"
        current_time = int(datetime.now().timestamp())
        
        with self.get_connection() as conn:
            cursor = conn.cursor()
            
            # 检查论文是否已存在
            cursor.execute("SELECT id FROM papers WHERE id = ?", (paper_id,))
            if cursor.fetchone():
                logger.info(f"论文已存在: {paper_info['title'][:50]}...")
                # 返回现有的paper_id和level_id
                cursor.execute("SELECT id FROM levels WHERE paper_id = ?", (paper_id,))
                existing_level = cursor.fetchone()
                if existing_level:
                    return paper_id, existing_level[0]
            
            # 插入论文
            authors_str = ', '.join(paper_info.get('authors', []))[:500]  # 限制长度
            cursor.execute("""
                INSERT OR REPLACE INTO papers 
                (id, subject_id, title, paper_author, paper_pub_ym, paper_citation_count, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                paper_id, subject_id, paper_info['title'][:500], 
                authors_str, str(paper_info.get('year', 2023)), 
                str(paper_info.get('citation_count', 0)), current_time, current_time
            ))
            
            # 插入关卡
            cursor.execute("""
                INSERT OR REPLACE INTO levels 
                (id, paper_id, name, pass_condition, meta_json, x, y, created_at, updated_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                level_id, paper_id, f"{paper_info['title'][:50]}...概念理解关卡", 
                "完成所有概念问题", json.dumps({"concepts_count": 5}), 
                0, 0, current_time, current_time
            ))
            
            conn.commit()
            logger.info(f"插入论文和关卡: {paper_info['title'][:50]}...")
            return paper_id, level_id

    def insert_question(self, level_id: str, question_data: Dict, concept_name: str) -> Tuple[str, str]:
        """插入两个问题到数据库：引入题和概念题"""
        current_time = int(datetime.now().timestamp())
        
        # 生成两个问题的ID
        lead_in_question_id = str(uuid.uuid4())
        concept_question_id = str(uuid.uuid4())
        
        # 构建引入题内容JSON
        lead_in_content_json = {
            "type": "analogical_lead_in",
            "concept_name": concept_name,
            "question": question_data['lead_in_question'],
            "options": question_data['lead_in_options']
        }
        
        # 构建概念题内容JSON
        concept_content_json = {
            "type": "conceptual_question",
            "concept_name": concept_name,
            "concept_explanation": question_data['concept_explanation'],
            "question": question_data['concept_question'],
            "options": question_data['concept_options']
        }
        
        # 构建答案JSON（两个题的正确答案应该是一致的）
        answer_json = {
            "correct_option": question_data['correct_option'],
            "explanation": question_data['explanation']
        }
        
        with self.get_connection() as conn:
            cursor = conn.cursor()
            
            # 插入引入题
            cursor.execute("""
                INSERT INTO questions 
                (id, level_id, stem, content_json, answer_json, score, created_by, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                lead_in_question_id, level_id, f"{concept_name} - 引入题", 
                json.dumps(lead_in_content_json, ensure_ascii=False),
                json.dumps(answer_json, ensure_ascii=False),
                5,  # 引入题分数
                "system_generated", current_time
            ))
            
            # 插入概念题
            cursor.execute("""
                INSERT INTO questions 
                (id, level_id, stem, content_json, answer_json, score, created_by, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                concept_question_id, level_id, f"{concept_name} - 概念题", 
                json.dumps(concept_content_json, ensure_ascii=False),
                json.dumps(answer_json, ensure_ascii=False),
                5,  # 概念题分数
                "system_generated", current_time
            ))
            
            conn.commit()
            logger.info(f"插入两个问题: {concept_name} (引入题 + 概念题)")
            return lead_in_question_id, concept_question_id

def load_concept_json(file_path: str) -> Optional[Dict]:
    """加载概念JSON文件"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            data = json.load(f)
        return data
    except Exception as e:
        logger.error(f"加载概念文件失败 {file_path}: {e}")
        return None

def process_all_concept_files():
    """处理所有概念文件"""
    papers_dir = Path(project_root) / "papers"
    
    if not papers_dir.exists():
        print("❌ papers目录不存在")
        return
    
    # 获取所有概念JSON文件
    concept_files = list(papers_dir.glob("*.concepts.json"))
    
    if not concept_files:
        print("❌ 未找到任何.concepts.json文件")
        return
    
    print(f"📁 找到 {len(concept_files)} 个概念文件")
    print(f"💡 预期生成: {len(concept_files) * 5 * 2} 道问题（{len(concept_files)} 文件 × 5 概念 × 2 问题）")
    
    # 初始化组件
    try:
        question_generator = AnalogicalQuestionGenerator()
        db_manager = DatabaseManager(os.path.join(project_root, 'sqlite', 'paperplay.db'))
        print("✅ 系统组件初始化成功")
    except Exception as e:
        print(f"❌ 系统初始化失败: {e}")
        return
    
    # 确保学科存在
    subject_id = db_manager.ensure_subject_exists()
    
    # 处理结果统计
    results = {
        "total_files": len(concept_files),
        "success_files": 0,
        "failed_files": 0,
        "total_questions": 0,
        "success_questions": 0,
        "failed_questions": 0,
        "details": []
    }
    
    for i, concept_file in enumerate(concept_files, 1):
        arxiv_id = concept_file.stem.replace('.concepts', '')
        print(f"\n🔍 [{i}/{len(concept_files)}] 处理: {arxiv_id}")
        print("-" * 50)
        
        try:
            # 加载概念数据
            concept_data = load_concept_json(str(concept_file))
            if not concept_data:
                results["failed_files"] += 1
                results["details"].append({
                    "file": str(concept_file),
                    "status": "failed",
                    "message": "概念文件加载失败"
                })
                print(f"❌ 概念文件加载失败: {arxiv_id}")
                continue
            
            paper_info = concept_data['paper_info']
            concepts = concept_data['concepts']
            
            print(f"📄 论文: {paper_info['title'][:50]}...")
            print(f"🧠 概念数量: {len(concepts)}")
            
            # 插入论文和关卡
            paper_id, level_id = db_manager.insert_paper_and_level(subject_id, paper_info)
            
            # 为每个概念生成问题
            file_success_count = 0
            file_failed_count = 0
            
            for j, concept in enumerate(concepts, 1):
                concept_name = concept.get('name', f'概念{j}')
                print(f"\n  🎯 [{j}/5] 生成问题: {concept_name}")
                
                try:
                    # 生成问题
                    question_data = question_generator.generate_question_for_concept(
                        paper_info['title'], concept
                    )
                    
                    if question_data:
                        # 存储到数据库（现在返回两个问题ID）
                        lead_in_id, concept_id = db_manager.insert_question(level_id, question_data, concept_name)
                        file_success_count += 2  # 因为生成了两个问题
                        results["success_questions"] += 2  # 因为生成了两个问题
                        print(f"  ✅ 问题生成成功: {concept_name}")
                        print(f"     引入题ID: {lead_in_id[:8]}... | {question_data['lead_in_question'][:30]}...")
                        print(f"     概念题ID: {concept_id[:8]}... | {question_data['concept_question'][:30]}...")
                        print(f"     正确答案: {question_data['correct_option']}")
                    else:
                        file_failed_count += 1
                        results["failed_questions"] += 1
                        print(f"  ❌ 问题生成失败: {concept_name}")
                        
                    results["total_questions"] += 1
                    
                except Exception as e:
                    file_failed_count += 1
                    results["failed_questions"] += 1
                    results["total_questions"] += 1
                    print(f"  ❌ 处理概念异常: {concept_name} - {e}")
            
            # 文件处理结果
            if file_failed_count == 0:
                results["success_files"] += 1
                results["details"].append({
                    "file": str(concept_file),
                    "status": "success",
                    "message": f"成功生成 {file_success_count} 道问题",
                    "questions_count": file_success_count
                })
                print(f"\n✅ 文件处理完成: {arxiv_id} ({file_success_count}道问题)")
            else:
                results["failed_files"] += 1
                results["details"].append({
                    "file": str(concept_file),
                    "status": "partial",
                    "message": f"部分成功: {file_success_count}成功, {file_failed_count}失败",
                    "questions_count": file_success_count
                })
                print(f"\n⚠️ 文件部分完成: {arxiv_id} ({file_success_count}成功, {file_failed_count}失败)")
                
        except Exception as e:
            results["failed_files"] += 1
            results["details"].append({
                "file": str(concept_file),
                "status": "failed",
                "message": str(e)
            })
            print(f"\n❌ 文件处理异常: {arxiv_id} - {e}")
    
    # 显示最终结果
    print("\n" + "=" * 60)
    print("🎯 问题生成结果摘要")
    print("=" * 60)
    print(f"📁 文件处理: {results['success_files']}成功 / {results['failed_files']}失败 / {results['total_files']}总计")
    print(f"❓ 问题生成: {results['success_questions']}成功 / {results['failed_questions']}失败 / {results['total_questions']}总计")
    print(f"📊 成功率: {results['success_questions']/results['total_questions']*100:.1f}%" if results['total_questions'] > 0 else "📊 成功率: 0%")
    print(f"💡 注意: 每个概念生成2道问题（引入题+概念题）")
    
    if results['success_questions'] > 0:
        print(f"\n✅ 成功生成的问题:")
        for detail in results['details']:
            if detail['status'] in ['success', 'partial'] and detail.get('questions_count', 0) > 0:
                arxiv_id = Path(detail['file']).stem.replace('.concepts', '')
                concepts_count = detail['questions_count'] // 2  # 每个概念2道题
                print(f"  - {arxiv_id}: {detail['questions_count']} 道问题（{concepts_count} 个概念 × 2）")
    
    print(f"\n💾 所有问题已存储到数据库: {db_manager.db_path}")
    print("🎉 处理完成！")

if __name__ == "__main__":
    print("🚀 类比引入式问题生成系统")
    print("从概念JSON文件生成教育问题并存储到数据库")
    print("=" * 60)
    
    process_all_concept_files()
