#!/usr/bin/env python3
"""
改进的题目生成Agent - 基于分层学习理论和认知负荷理论
"""

import os
import sys
import json
import logging
from typing import Dict, List, Any, Optional, Tuple
from langchain.tools import tool
from langgraph.prebuilt import create_react_agent
from langchain_community.chat_models import ChatTongyi

# 导入工具类
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
try:
    from utils import DatabaseManager, PDFTextExtractor
except ImportError:
    print("⚠️  未找到utils模块，某些功能可能受限")
    DatabaseManager = None
    PDFTextExtractor = None

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# 加载环境变量
try:
    from dotenv import load_dotenv
    load_dotenv()
    print("✅ 已加载.env文件中的环境变量")
except ImportError:
    print("⚠️  未安装python-dotenv包，无法加载.env文件")

# 检查API密钥
api_key = os.getenv("DASHSCOPE_API_KEY")
if not api_key:
    print("❌ 错误: 未找到DASHSCOPE_API_KEY环境变量")
    print("   请在.env文件中添加: DASHSCOPE_API_KEY=your-api-key")
else:
    print("✅ 找到DASHSCOPE_API_KEY")

# ========== 第一层：知识结构分析层 ==========

@tool
def extract_core_concepts(paper_content: str, paper_title: str) -> str:
    """
    从论文中提取核心概念及其定义
    
    Args:
        paper_content (str): 论文完整内容
        paper_title (str): 论文标题
        
    Returns:
        str: 核心概念JSON格式 {"concepts": [{"name": "", "definition": "", "importance": 1-5}]}
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.3)
        
    prompt = f"""
分析以下论文，提取5-10个最核心的概念：

论文标题：{paper_title}
论文内容：
{paper_content[:3000]}

要求：
1. 提取最基础、最核心的概念
2. 给出简洁的定义（不超过100字）
3. 按重要性排序（1-5分，5分最重要）
4. 确保概念之间有逻辑关联

返回JSON格式：
{{
    "concepts": [
        {{
            "name": "概念名称",
            "definition": "概念定义", 
            "importance": 5,
            "category": "基础理论/方法技术/应用实践"
        }}
    ]
}}
"""
    
    response = model.invoke(prompt)
    return response.content

@tool
def analyze_knowledge_dependencies(concepts_json: str) -> str:
    """
    分析知识点之间的依赖关系
    
    Args:
        concepts_json (str): 核心概念JSON
        
    Returns:
        str: 依赖关系JSON格式
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.3)
    
    prompt = f"""
分析以下概念之间的学习依赖关系：

概念列表：
{concepts_json}

要求：
1. 确定概念的学习顺序（哪些是前置概念）
2. 标识概念之间的强依赖和弱依赖关系
3. 将概念分为不同的学习层级

返回JSON格式：
{{
    "learning_sequence": ["概念1", "概念2", ...],
    "dependencies": [
        {{
            "prerequisite": "前置概念",
            "dependent": "依赖概念", 
            "strength": "strong/weak"
        }}
    ],
    "levels": {{
        "foundation": ["基础概念"],
        "intermediate": ["中级概念"],
        "advanced": ["高级概念"]
    }}
}}
"""
        
    response = model.invoke(prompt)
    return response.content

@tool
def classify_cognitive_levels(concepts_json: str, dependencies_json: str) -> str:
    """
    按认知层次对概念进行分类（布鲁姆分类法）
    
    Args:
        concepts_json (str): 核心概念JSON
        dependencies_json (str): 依赖关系JSON
        
    Returns:
        str: 认知层次分类JSON
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.3)
        
    prompt = f"""
基于布鲁姆认知分类法，对概念进行层次分类：

概念：{concepts_json}
依赖关系：{dependencies_json}

认知层次：
1. 记忆（Remember）- 基础事实和定义
2. 理解（Understand）- 概念解释和关系
3. 应用（Apply）- 实际问题解决
4. 分析（Analyze）- 复杂关系分析
5. 综合（Synthesize）- 创新整合
6. 评价（Evaluate）- 批判性思维

返回JSON格式：
{{
    "cognitive_mapping": {{
        "remember": ["基础概念定义"],
        "understand": ["概念原理解释"],
        "apply": ["实际应用场景"],
        "analyze": ["复杂关系分析"],
        "synthesize": ["创新整合"], 
        "evaluate": ["批判评价"]
    }}
}}
"""
        
    response = model.invoke(prompt)
    return response.content

# ========== 第二层：概念锚定层 ==========

@tool
def map_existing_knowledge(concept_name: str, concept_definition: str) -> str:
    """
    将新概念映射到学习者已有的常识知识
    
    Args:
        concept_name (str): 概念名称
        concept_definition (str): 概念定义
        
    Returns:
        str: 知识映射JSON
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.7)
    
    prompt = f"""
将学术概念连接到日常生活常识：

概念：{concept_name}
定义：{concept_definition}

要求：
1. 找出3-5个日常生活中的相似例子
2. 建立从熟悉到陌生的知识桥梁
3. 用通俗语言重新解释概念

返回JSON格式：
{{
    "everyday_examples": ["日常例子1", "日常例子2"],
    "analogies": ["类比1", "类比2"],
    "simplified_explanation": "通俗化解释",
    "knowledge_bridge": "从已知到未知的连接过程"
}}
"""
    
    response = model.invoke(prompt)
    return response.content

@tool
def generate_concept_analogies(concept_name: str, target_audience: str = "本科生") -> str:
    """
    为复杂概念生成类比和隐喻
    
    Args:
        concept_name (str): 概念名称
        target_audience (str): 目标受众
        
    Returns:
        str: 类比隐喻JSON
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.8)
        
    prompt = f"""
为复杂概念创造生动的类比和隐喻：

概念：{concept_name}
目标受众：{target_audience}

要求：
1. 创造2-3个生动的类比
2. 确保类比准确且易于理解
3. 解释类比的对应关系

返回JSON格式：
{{
    "analogies": [
        {{
            "analogy": "类比描述",
            "explanation": "类比解释",
            "correspondence": "对应关系说明"
        }}
    ],
    "metaphors": ["隐喻1", "隐喻2"]
}}
"""
        
    response = model.invoke(prompt)
    return response.content

# ========== 第三层：分层题目生成层 ==========

@tool
def generate_memory_level_questions(concept_name: str, definition: str, knowledge_mapping: str) -> str:
    """
    生成记忆层次题目（基础概念识别）
    
    Args:
        concept_name (str): 概念名称
        definition (str): 概念定义
        knowledge_mapping (str): 知识映射JSON
        
    Returns:
        str: 记忆层题目JSON
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.5)
    
    prompt = f"""
基于概念生成记忆层次的学习题目：

概念：{concept_name}
定义：{definition}
知识映射：{knowledge_mapping}

记忆层次要求：
- 测试基础事实记忆
- 概念定义识别
- 关键术语理解
- 使用简单直接的语言

生成2道不同类型的题目：
1. 选择题（概念定义匹配）
2. 填空题（关键词填空）

返回JSON格式：
{{
    "questions": [
        {{
            "level": "memory",
            "type": "multiple_choice",
            "stem": "题干",
            "options": ["A. 选项1", "B. 选项2", "C. 选项3", "D. 选项4"],
            "correct_answer": "A",
            "explanation": "答案解释",
            "cognitive_focus": "测试的认知要点"
        }}
    ]
}}
"""
    
    response = model.invoke(prompt)
    return response.content

@tool
def generate_understanding_level_questions(concept_name: str, analogies: str, dependencies: str) -> str:
    """
    生成理解层次题目（概念关系和原理）
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.6)
    
    prompt = f"""
生成理解层次的题目：

概念：{concept_name}
类比：{analogies}
依赖关系：{dependencies}

理解层次要求：
- 测试概念之间的关系
- 原理的解释能力
- 使用类比帮助理解
- 包含"为什么"类型的问题

生成2-3道题目，包含：
1. 解释题（概念原理解释）
2. 关系题（概念间关系）
3. 类比题（使用类比理解）

返回JSON格式：
{{
    "questions": [
        {{
            "level": "understanding",
            "type": "short_answer",
            "stem": "题干",
            "correct_answer": "正确答案",
            "explanation": "答案解释",
            "cognitive_focus": "测试的认知要点"
        }}
    ]
}}
"""
    
    response = model.invoke(prompt)
    return response.content

@tool
def generate_application_level_questions(concept_name: str, real_world_examples: str) -> str:
    """
    生成应用层次题目（实际问题解决）
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.6)
    
    prompt = f"""
生成应用层次的题目：

概念：{concept_name}
实际例子：{real_world_examples}

应用层次要求：
- 测试在实际场景中的应用能力
- 问题解决技能
- 概念的实际运用
- 情境化的学习

生成2道应用题：
1. 案例分析题
2. 实际问题解决题

返回JSON格式：
{{
    "questions": [
        {{
            "level": "application",
            "type": "case_study",
            "stem": "题干",
            "correct_answer": "正确答案",
            "explanation": "答案解释",
            "cognitive_focus": "测试的认知要点"
        }}
    ]
}}
"""
    
    response = model.invoke(prompt)
    return response.content

# ========== 第四层：认知负荷优化层 ==========

@tool
def simplify_question_language(question_json: str, target_level: str = "undergraduate") -> str:
    """
    简化题目语言，降低认知负荷
    
    Args:
        question_json (str): 原始题目JSON
        target_level (str): 目标水平
        
    Returns:
        str: 优化后的题目JSON
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.3)
    
    prompt = f"""
优化题目语言，降低认知负荷：

原始题目：{question_json}
目标水平：{target_level}

优化要求：
1. 使用简单直白的语言
2. 避免复杂的句式结构
3. 确保每题只考察一个知识点
4. 提供必要的背景信息
5. 去除多余的修饰词

返回优化后的题目JSON，保持相同的结构。
"""
    
    response = model.invoke(prompt)
    return response.content

@tool
def ensure_single_concept_focus(question_json: str) -> str:
    """
    确保题目聚焦单一知识点
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.3)
    
    prompt = f"""
检查并优化题目，确保单一知识点聚焦：

题目：{question_json}

检查要求：
1. 题目是否只测试一个核心概念
2. 是否有多余的干扰信息
3. 选项是否清晰区分
4. 认知要求是否一致

返回优化后的题目JSON，如果没有问题则保持原样。
"""
        
    response = model.invoke(prompt)
    return response.content

# ========== 第五层：智能反馈层 ==========

@tool
def diagnose_error_types(question_json: str, common_mistakes: List[str]) -> str:
    """
    预测学习者可能的错误类型
    
    Args:
        question_json (str): 题目JSON
        common_mistakes (List[str]): 常见错误类型
        
    Returns:
        str: 错误诊断JSON
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.5)
    
    prompt = f"""
分析学习者在此题目上可能犯的错误：

题目：{question_json}
常见错误类型：{common_mistakes}

分析要求：
1. 预测3-5种可能的错误类型
2. 分析错误的认知根源
3. 为每种错误设计针对性反馈

返回JSON格式：
{{
    "error_predictions": [
        {{
            "error_type": "错误类型",
            "description": "错误描述", 
            "cognitive_cause": "认知原因",
            "frequency": "high/medium/low",
            "remediation_strategy": "补救策略"
        }}
    ]
}}
"""
    
    response = model.invoke(prompt)
    return response.content

@tool
def generate_personalized_feedback(student_answer: str, correct_answer: str, error_diagnosis: str) -> str:
    """
    生成个性化的学习指导反馈
    
    Args:
        student_answer (str): 学生答案
        correct_answer (str): 正确答案
        error_diagnosis (str): 错误诊断
        
    Returns:
        str: 个性化反馈JSON
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.6)
    
    prompt = f"""
生成个性化的学习指导反馈：

学生答案：{student_answer}
正确答案：{correct_answer}  
错误诊断：{error_diagnosis}

反馈要求：
1. 先肯定答对的部分（如果有）
2. 指出具体的错误点
3. 解释为什么会有这个错误
4. 提供学习建议和资源
5. 鼓励继续学习

返回JSON格式：
{{
    "positive_reinforcement": "肯定的部分",
    "error_identification": "错误识别",
    "explanation": "错误原因解释",
    "correct_reasoning": "正确思路引导", 
    "learning_suggestions": ["学习建议1", "学习建议2"],
    "next_steps": "下一步学习方向",
    "encouragement": "鼓励话语"
}}
"""
    
    response = model.invoke(prompt)
    return response.content

@tool 
def recommend_learning_path(current_level: str, mastered_concepts: List[str], target_concepts: List[str]) -> str:
    """
    推荐个性化学习路径
    """
    model = ChatTongyi(model="qwen-max-latest", temperature=0.5)
    
    prompt = f"""
为学习者推荐个性化学习路径：

当前水平：{current_level}
已掌握概念：{mastered_concepts}
目标概念：{target_concepts}

推荐要求：
1. 分析知识差距
2. 制定循序渐进的学习计划
3. 推荐学习资源和方法
4. 设置学习里程碑

返回JSON格式：
{{
    "knowledge_gap": "知识差距分析",
    "learning_plan": [
        {{
            "step": 1,
            "concept": "学习概念",
            "duration": "建议时间",
            "resources": ["资源1", "资源2"]
        }}
    ],
    "milestones": ["里程碑1", "里程碑2"],
    "study_tips": ["学习建议1", "学习建议2"]
}}
"""
    
    response = model.invoke(prompt)
    return response.content

# ========== 基础工具（兼容性） ==========

@tool
def save_question_to_database(level_id: str, question_json: str) -> str:
    """
    将生成的题目保存到数据库
    """
    try:
        if DatabaseManager is None:
            return "⚠️ 数据库管理器不可用，题目未保存"
        
        question_data = json.loads(question_json)
        db_manager = DatabaseManager()
        
        # 处理新格式的题目数据
        if "questions" in question_data:
            # 新格式：包含多个题目
            results = []
            for q in question_data["questions"]:
                question_id = db_manager.insert_question(
                    level_id=level_id,
                    stem=q["stem"],
                    content_json=q,
                    answer_json={"correct_answer": q.get("correct_answer"), "explanation": q.get("explanation")},
                    score=1,
                    created_by="enhanced_agent"
                )
                results.append(f"题目ID: {question_id}")
            return f"✅ 保存成功！{', '.join(results)}"
        else:
            # 旧格式兼容
            question_id = db_manager.insert_question(
                level_id=level_id,
                stem=question_data.get("stem", "题目"),
                content_json=question_data,
                answer_json=question_data.get("answer_json", {}),
                score=1,
                created_by="enhanced_agent"
            )
            return f"✅ 题目保存成功！题目ID: {question_id}"
        
    except Exception as e:
        logger.error(f"保存题目失败: {e}")
        return f"❌ 保存题目失败: {str(e)}"

@tool
def get_question_generation_stats() -> str:
    """
    获取题目生成统计信息
    """
    try:
        if DatabaseManager is None:
            return "⚠️ 数据库管理器不可用，无法获取统计信息"
        
        db_manager = DatabaseManager()
        with db_manager.get_connection() as conn:
            cursor = conn.execute("""
                SELECT 
                    COUNT(*) as total_questions,
                    COUNT(DISTINCT level_id) as levels_with_questions
                FROM questions 
                WHERE created_by = 'enhanced_agent'
            """)
            stats = cursor.fetchone()
        
        return f"""
📊 增强Agent题目生成统计信息

总体统计:
- 增强Agent生成题目总数: {stats['total_questions']} 题
- 涉及关卡数: {stats['levels_with_questions']} 个

生成引擎: enhanced_agent（基于认知科学）
"""
        
    except Exception as e:
        logger.error(f"获取统计信息失败: {e}")
        return f"❌ 获取统计信息失败: {str(e)}"

# ========== Agent创建函数 ==========

def create_enhanced_question_generation_agent():
    """创建增强的题目生成Agent"""

    # 初始化LLM
    model = ChatTongyi(
        model="qwen-max-latest",
        top_p=0.8,
        streaming=True,
        temperature=0.7,
        max_retries=3
    )
    
    # 定义分层工具列表
    tools = [
        # 知识结构分析层
        extract_core_concepts,
        analyze_knowledge_dependencies, 
        classify_cognitive_levels,
        
        # 概念锚定层
        map_existing_knowledge,
        generate_concept_analogies,
        
        # 分层题目生成层
        generate_memory_level_questions,
        generate_understanding_level_questions,
        generate_application_level_questions,
        
        # 认知负荷优化层
        simplify_question_language,
        ensure_single_concept_focus,
        
        # 智能反馈层
        diagnose_error_types,
        generate_personalized_feedback,
        recommend_learning_path,
        
        # 基础工具
        save_question_to_database,
        get_question_generation_stats
    ]
    
    # 系统提示
    system_prompt = """你是一个基于认知科学的智能题目生成助手。

你的核心理念：
1. 分层学习理论：从基础概念到高级应用，层层递进
2. 认知负荷理论：简化语言，聚焦单一知识点，提供清晰反馈
3. 概念锚定理论：将新知识与已有知识建立连接

工作流程：
1. 首先分析论文的知识结构和概念依赖关系
2. 为复杂概念建立与常识的连接
3. 按认知层次生成不同难度的题目
4. 优化语言表达，降低认知负荷
5. 设计智能反馈机制，帮助学习者理解错误

始终记住：好的题目不仅测试知识，更要促进理解和学习。"""
    
    # 创建Agent
    agent = create_react_agent(
        model=model,
        tools=tools,
        prompt=system_prompt
    )
    
    return agent
    
def run_enhanced_question_generation_agent(message: str, thread_id: str = "1") -> dict:
    """
    运行增强的题目生成Agent
    
    Args:
        message (str): 用户消息
        thread_id (str): 会话ID
        
    Returns:
        dict: 处理结果
    """
    try:
        agent = create_enhanced_question_generation_agent()
        
        # 执行Agent
        config = {"configurable": {"thread_id": thread_id}}
        
        responses = []
        for chunk in agent.stream({"messages": [("human", message)]}, config=config):
            if "agent" in chunk:
                if chunk["agent"].get("messages"):
                    for msg in chunk["agent"]["messages"]:
                        if hasattr(msg, 'content'):
                            responses.append(msg.content)
            elif "tools" in chunk:
                for tool_name, tool_result in chunk["tools"].items():
                    responses.append(f"🔧 {tool_name}: {tool_result}")
        
        return {"status": "success", "responses": responses, "message": "增强题目生成完成"}
        
    except Exception as e:
        logger.error(f"运行增强Agent失败: {e}")
        return {"status": "error", "message": str(e), "responses": []}

# ========== 便捷函数 ==========

def generate_enhanced_questions_for_paper(paper_content: str, paper_title: str) -> dict:
    """
    为指定论文生成增强题目的便捷方法
    """
    message = f"""
请使用增强的题目生成流程为以下论文生成题目：

论文标题：{paper_title}
论文内容：{paper_content[:1000]}...

请按照以下步骤：
1. 提取核心概念
2. 分析知识依赖关系
3. 为每个概念生成不同认知层次的题目
4. 优化题目语言和结构
"""
    return run_enhanced_question_generation_agent(message, thread_id=f"paper_{hash(paper_title)}")

if __name__ == "__main__":
    print("🤖 增强题目生成Agent 启动")
    print("=" * 60)
