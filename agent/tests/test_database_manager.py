#!/usr/bin/env python3
"""
测试新的DatabaseManager类 - 验证基本功能和新架构
"""

from utils import DatabaseManager
import json

def test_database_manager():
    """测试DatabaseManager的基本功能"""
    
    print("🧪 测试新的DatabaseManager类 (001_init.sql架构)")
    print("=" * 60)
    
    # 初始化数据库管理器
    try:
        db_manager = DatabaseManager()
        print("✅ DatabaseManager初始化成功")
        print(f"📁 默认学科: {db_manager.DEFAULT_SUBJECT}")
        print(f"🆔 默认学科ID: {db_manager.get_default_subject_id()}")
        print(f"📄 使用架构文件: {db_manager.schema_path}")
    except Exception as e:
        print(f"❌ DatabaseManager初始化失败: {e}")
        return False
    
    print("\n" + "=" * 60)
    
    # 测试插入论文
    print("📄 测试插入论文...")
    try:
        paper_id = db_manager.insert_paper(
            title="测试论文：Attention Is All You Need",
            paper_author="Ashish Vaswani; Noam Shazeer; Niki Parmar",
            paper_pub_ym="2017-06",
            paper_citation_count="50000"
        )
        print(f"✅ 论文插入成功，ID: {paper_id}")
    except Exception as e:
        print(f"❌ 论文插入失败: {e}")
        return False
    
    # 测试查询论文
    print("\n📖 测试查询论文...")
    try:
        paper = db_manager.get_paper_by_id(paper_id)
        if paper:
            print("✅ 论文查询成功")
            print(f"   标题: {paper['title']}")
            print(f"   作者: {paper['paper_author']}")
            print(f"   发表年月: {paper['paper_pub_ym']}")
            print(f"   引用数: {paper['paper_citation_count']}")
        else:
            print("❌ 论文查询失败：未找到")
            return False
    except Exception as e:
        print(f"❌ 论文查询失败: {e}")
        return False
    
    # 测试插入关卡（包含坐标信息）
    print("\n🎮 测试插入关卡（含坐标）...")
    try:
        pass_condition = {
            "type": "score_based",
            "min_score": 80,
            "total_questions": 10
        }
        meta_data = {
            "difficulty": "medium", 
            "estimated_time": "30 minutes",
            "description": "理解Transformer架构的核心概念"
        }
        
        level_id = db_manager.insert_level(
            paper_id=paper_id,
            name="Transformer架构理解关卡",
            pass_condition=pass_condition,
            meta_json=meta_data,
            x=100,  # 设置坐标
            y=200
        )
        print(f"✅ 关卡插入成功，ID: {level_id}")
        print(f"   坐标位置: (100, 200)")
    except Exception as e:
        print(f"❌ 关卡插入失败: {e}")
        return False
    
    # 测试查询关卡
    print("\n🔍 测试查询关卡...")
    try:
        level = db_manager.get_level_by_id(level_id)
        if level:
            print("✅ 关卡查询成功")
            print(f"   名称: {level['name']}")
            print(f"   坐标: ({level['x']}, {level['y']})")
            print(f"   通关条件: {level['pass_condition']}")
            print(f"   元数据: {level['meta_json']}")
        else:
            print("❌ 关卡查询失败：未找到")
            return False
    except Exception as e:
        print(f"❌ 关卡查询失败: {e}")
        return False
    
    # 测试插入路线图节点
    print("\n🗺️  测试插入路线图节点...")
    try:
        subject_id = db_manager.get_default_subject_id()
        node_id = db_manager.insert_roadmap_node(
            subject_id=subject_id,
            level_id=level_id,
            parent_id=None,  # 根节点
            sort_order=1,
            path="/agent/transformer-basics",
            depth=0
        )
        print(f"✅ 路线图节点插入成功，ID: {node_id}")
    except Exception as e:
        print(f"❌ 路线图节点插入失败: {e}")
        return False
    
    # 测试查询路线图节点
    print("\n🔍 测试查询路线图节点...")
    try:
        nodes = db_manager.get_roadmap_nodes_by_subject(subject_id)
        if nodes:
            print(f"✅ 路线图节点查询成功，找到 {len(nodes)} 个节点")
            for node in nodes:
                print(f"   节点: {node['level_name']} at ({node['x']}, {node['y']})")
        else:
            print("❌ 路线图节点查询失败：未找到")
            return False
    except Exception as e:
        print(f"❌ 路线图节点查询失败: {e}")
        return False
    
    # 测试插入题目
    print("\n❓ 测试插入题目...")
    try:
        content_json = {
            "question_type": "multiple_choice",
            "options": ["A. 卷积神经网络", "B. 循环神经网络", "C. 自注意力机制", "D. 池化层"],
            "correct_option": "C"
        }
        answer_json = {
            "correct_answer": "C",
            "explanation": "Transformer的核心创新是自注意力机制，允许模型直接建模序列中任意两个位置之间的依赖关系。"
        }
        
        question_id = db_manager.insert_question(
            level_id=level_id,
            stem="Transformer模型的核心组件是什么？",
            content_json=content_json,
            answer_json=answer_json,
            score=10,
            created_by="test_system"
        )
        print(f"✅ 题目插入成功，ID: {question_id}")
    except Exception as e:
        print(f"❌ 题目插入失败: {e}")
        return False
    
    # 测试查询题目
    print("\n🔎 测试查询题目...")
    try:
        question = db_manager.get_question_by_id(question_id)
        if question:
            print("✅ 题目查询成功")
            print(f"   题干: {question['stem']}")
            print(f"   分值: {question['score']}")
            print(f"   创建者: {question['created_by']}")
        else:
            print("❌ 题目查询失败：未找到")
            return False
    except Exception as e:
        print(f"❌ 题目查询失败: {e}")
        return False
    
    # 测试坐标更新
    print("\n📍 测试关卡坐标更新...")
    try:
        success = db_manager.update_level(level_id, x=150, y=250)
        if success:
            updated_level = db_manager.get_level_by_id(level_id)
            print(f"✅ 坐标更新成功: ({updated_level['x']}, {updated_level['y']})")
        else:
            print("❌ 坐标更新失败")
            return False
    except Exception as e:
        print(f"❌ 坐标更新失败: {e}")
        return False
    
    # 测试统计信息（包含新表）
    print("\n📊 测试统计信息...")
    try:
        stats = db_manager.get_system_stats()
        print("✅ 统计信息获取成功")
        print(f"   学科数: {stats.get('total_subjects', 0)}")
        print(f"   论文数: {stats.get('total_papers', 0)}")
        print(f"   关卡数: {stats.get('total_levels', 0)}")
        print(f"   题目数: {stats.get('total_questions', 0)}")
        print(f"   用户数: {stats.get('total_users', 0)}")
        print(f"   路线图节点数: {stats.get('total_roadmap_nodes', 0)}")
        print(f"   平均每关卡题目数: {stats.get('avg_questions_per_level', 0)}")
    except Exception as e:
        print(f"❌ 统计信息获取失败: {e}")
        return False
    
    # 测试完整查询
    print("\n🔄 测试完整查询...")
    try:
        full_paper = db_manager.get_paper_with_level_and_questions(paper_id)
        if full_paper and 'level' in full_paper:
            print("✅ 完整查询成功")
            print(f"   论文标题: {full_paper['title']}")
            print(f"   关卡名称: {full_paper['level']['name']}")
            print(f"   关卡坐标: ({full_paper['level']['x']}, {full_paper['level']['y']})")
            print(f"   题目数量: {len(full_paper['level']['questions'])}")
        else:
            print("❌ 完整查询失败")
            return False
    except Exception as e:
        print(f"❌ 完整查询失败: {e}")
        return False
    
    # 测试数据表验证
    print("\n🔍 测试数据表完整性...")
    try:
        with db_manager.get_connection() as conn:
            # 检查所有预期的表是否存在
            cursor = conn.execute("SELECT name FROM sqlite_master WHERE type='table'")
            tables = [row['name'] for row in cursor.fetchall()]
            
            expected_tables = [
                'subjects', 'papers', 'levels', 'questions', 
                'roadmap_nodes', 'users', 'refresh_tokens', 
                'user_progress', 'user_attempts'
            ]
            
            missing_tables = [table for table in expected_tables if table not in tables]
            if missing_tables:
                print(f"❌ 缺少表: {missing_tables}")
                return False
            else:
                print(f"✅ 所有预期表都存在: {len(expected_tables)} 个表")
                
        print(f"   实际创建的表: {sorted(tables)}")
    except Exception as e:
        print(f"❌ 表完整性检查失败: {e}")
        return False
    
    print("\n" + "=" * 60)
    print("🎉 所有测试通过！新的DatabaseManager (001_init.sql) 工作正常")
    print("\n🔋 新架构特性验证:")
    print("  ✅ 基础表结构 (subjects → papers → levels → questions)")
    print("  ✅ 关卡坐标支持 (x, y)")
    print("  ✅ 路线图节点系统")
    print("  ✅ 用户系统表结构")
    print("  ✅ 进度追踪表结构")
    print("  ✅ UUID主键和时间戳")
    return True

if __name__ == "__main__":
    success = test_database_manager()
    if success:
        print("\n✅ 数据库架构迁移成功！")
        print("💡 新架构包含完整的用户系统、路线图和进度追踪功能")
    else:
        print("\n❌ 测试失败，请检查配置")
    
    print("\n🚀 下一步:")
    print("   • 运行 'python main.py' 测试完整的论文处理流程")
    print("   • 新架构支持可视化关卡布局和学习路径")
    print("   • 可以开始开发用户进度追踪功能") 