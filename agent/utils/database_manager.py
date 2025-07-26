import sqlite3
import json
import os
import uuid
import time
from typing import List, Dict, Any, Optional, Tuple
import logging

class DatabaseManager:
    """教育内容管理系统数据库管理类 - 处理SQLite数据库的所有操作"""
    
    # MVP版本：默认学科名称
    DEFAULT_SUBJECT = "agent"
    
    def __init__(self, db_path: str = "sqlite/paperplay.db"):
        self.db_path = db_path
        self.schema_path = "sqlite/001_init.sql"  # 更新为新的架构文件
        self.logger = logging.getLogger(__name__)
        self._subject_id = None  # 缓存默认学科ID
        self._init_database()
    
    def _init_database(self):
        """初始化数据库，创建表结构和默认数据"""
        try:
            # 确保目录存在
            os.makedirs(os.path.dirname(self.db_path), exist_ok=True)
            
            # 执行schema文件
            if os.path.exists(self.schema_path):
                with open(self.schema_path, 'r', encoding='utf-8') as f:
                    schema_sql = f.read()
                
                with sqlite3.connect(self.db_path) as conn:
                    conn.executescript(schema_sql)
                    conn.commit()
                self.logger.info("数据库初始化完成")
                
                # 创建默认学科
                self._ensure_default_subject()
            else:
                self.logger.error(f"Schema文件不存在: {self.schema_path}")
        except Exception as e:
            self.logger.error(f"数据库初始化失败: {e}")
            raise
    
    def _ensure_default_subject(self):
        """确保默认学科存在"""
        try:
            with self.get_connection() as conn:
                # 检查是否已存在默认学科
                cursor = conn.execute("SELECT id FROM subjects WHERE name = ?", (self.DEFAULT_SUBJECT,))
                row = cursor.fetchone()
                
                if row:
                    self._subject_id = row['id']
                    self.logger.info(f"找到默认学科: {self.DEFAULT_SUBJECT} (ID: {self._subject_id})")
                else:
                    # 创建默认学科
                    subject_id = str(uuid.uuid4())
                    current_time = int(time.time())
                    
                    conn.execute("""
                        INSERT INTO subjects (id, name, description, created_at, updated_at)
                        VALUES (?, ?, ?, ?, ?)
                    """, (subject_id, self.DEFAULT_SUBJECT, "AI Agent相关论文学科", current_time, current_time))
                    
                    conn.commit()
                    self._subject_id = subject_id
                    self.logger.info(f"创建默认学科: {self.DEFAULT_SUBJECT} (ID: {self._subject_id})")
        except Exception as e:
            self.logger.error(f"创建默认学科失败: {e}")
            raise
    
    def get_connection(self) -> sqlite3.Connection:
        """获取数据库连接"""
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row  # 返回字典式的行对象
        return conn
    
    def _generate_id(self) -> str:
        """生成UUID"""
        return str(uuid.uuid4())
    
    def _get_current_timestamp(self) -> int:
        """获取当前时间戳"""
        return int(time.time())
    
    # ========== 学科相关操作 (MVP版本简化) ==========
    
    def get_default_subject_id(self) -> str:
        """获取默认学科ID"""
        if not self._subject_id:
            self._ensure_default_subject()
        return self._subject_id
    
    def get_subject_by_id(self, subject_id: str) -> Optional[Dict]:
        """根据ID查询学科"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM subjects WHERE id = ?", (subject_id,))
                row = cursor.fetchone()
                return dict(row) if row else None
        except Exception as e:
            self.logger.error(f"查询学科失败: {e}")
            return None
    
    # ========== 论文相关操作 ==========
    
    def insert_paper(self, title: str, paper_author: str, paper_pub_ym: str, 
                    paper_citation_count: str = "0") -> str:
        """插入论文记录 (MVP版本：自动归属到默认学科)"""
        try:
            paper_id = self._generate_id()
            current_time = self._get_current_timestamp()
            subject_id = self.get_default_subject_id()
            
            with self.get_connection() as conn:
                conn.execute("""
                    INSERT INTO papers (id, subject_id, title, paper_author, paper_pub_ym, 
                                      paper_citation_count, created_at, updated_at)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """, (paper_id, subject_id, title, paper_author, paper_pub_ym, 
                     paper_citation_count, current_time, current_time))
                
                conn.commit()
                self.logger.info(f"成功插入论文: {title}, ID: {paper_id}")
                return paper_id
        except Exception as e:
            self.logger.error(f"插入论文失败: {e}")
            raise
    
    def get_paper_by_id(self, paper_id: str) -> Optional[Dict]:
        """根据ID查询论文"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM papers WHERE id = ?", (paper_id,))
                row = cursor.fetchone()
                return dict(row) if row else None
        except Exception as e:
            self.logger.error(f"查询论文失败: {e}")
            return None
    
    def get_paper_by_title(self, title: str) -> Optional[Dict]:
        """根据标题查询论文"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM papers WHERE title = ?", (title,))
                row = cursor.fetchone()
                return dict(row) if row else None
        except Exception as e:
            self.logger.error(f"查询论文失败: {e}")
            return None
    
    def get_papers_by_subject(self, subject_id: str = None) -> List[Dict]:
        """查询学科下的所有论文"""
        if not subject_id:
            subject_id = self.get_default_subject_id()
        
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM papers WHERE subject_id = ? ORDER BY created_at DESC", (subject_id,))
                return [dict(row) for row in cursor.fetchall()]
        except Exception as e:
            self.logger.error(f"查询论文列表失败: {e}")
            return []
    
    def update_paper(self, paper_id: str, **kwargs) -> bool:
        """更新论文信息"""
        if not kwargs:
            return False
        
        try:
            # 构建动态更新SQL
            set_clauses = []
            values = []
            
            for field, value in kwargs.items():
                if field in ['title', 'paper_author', 'paper_pub_ym', 'paper_citation_count']:
                    set_clauses.append(f"{field} = ?")
                    values.append(value)
            
            if not set_clauses:
                return False
            
            # 添加更新时间
            set_clauses.append("updated_at = ?")
            values.append(self._get_current_timestamp())
            values.append(paper_id)
            
            sql = f"UPDATE papers SET {', '.join(set_clauses)} WHERE id = ?"
            
            with self.get_connection() as conn:
                cursor = conn.execute(sql, values)
                conn.commit()
                
                if cursor.rowcount > 0:
                    self.logger.info(f"更新论文成功: {paper_id}")
                    return True
                else:
                    self.logger.warning(f"论文不存在: {paper_id}")
                    return False
        except Exception as e:
            self.logger.error(f"更新论文失败: {e}")
            return False
    
    def delete_paper(self, paper_id: str) -> bool:
        """删除论文 (级联删除关卡和题目)"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("DELETE FROM papers WHERE id = ?", (paper_id,))
                conn.commit()
                
                if cursor.rowcount > 0:
                    self.logger.info(f"删除论文成功: {paper_id}")
                    return True
                else:
                    self.logger.warning(f"论文不存在: {paper_id}")
                    return False
        except Exception as e:
            self.logger.error(f"删除论文失败: {e}")
            return False
    
    # ========== 关卡相关操作 ==========
    
    def insert_level(self, paper_id: str, name: str, pass_condition: Dict, 
                    meta_json: Dict = None, x: int = 0, y: int = 0) -> str:
        """插入关卡记录 (支持坐标位置)"""
        try:
            level_id = self._generate_id()
            current_time = self._get_current_timestamp()
            
            with self.get_connection() as conn:
                conn.execute("""
                    INSERT INTO levels (id, paper_id, name, pass_condition, meta_json, x, y, created_at, updated_at)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (level_id, paper_id, name, json.dumps(pass_condition, ensure_ascii=False),
                     json.dumps(meta_json, ensure_ascii=False) if meta_json else None,
                     x, y, current_time, current_time))
                
                conn.commit()
                self.logger.info(f"成功插入关卡: {name}, ID: {level_id}")
                return level_id
        except Exception as e:
            self.logger.error(f"插入关卡失败: {e}")
            raise
    
    def get_level_by_id(self, level_id: str) -> Optional[Dict]:
        """根据ID查询关卡"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM levels WHERE id = ?", (level_id,))
                row = cursor.fetchone()
                if row:
                    result = dict(row)
                    # 解析JSON字段
                    result['pass_condition'] = json.loads(result['pass_condition'])
                    if result['meta_json']:
                        result['meta_json'] = json.loads(result['meta_json'])
                    return result
                return None
        except Exception as e:
            self.logger.error(f"查询关卡失败: {e}")
            return None
    
    def get_level_by_paper_id(self, paper_id: str) -> Optional[Dict]:
        """根据论文ID查询关卡 (一对一关系)"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM levels WHERE paper_id = ?", (paper_id,))
                row = cursor.fetchone()
                if row:
                    result = dict(row)
                    # 解析JSON字段
                    result['pass_condition'] = json.loads(result['pass_condition'])
                    if result['meta_json']:
                        result['meta_json'] = json.loads(result['meta_json'])
                    return result
                return None
        except Exception as e:
            self.logger.error(f"查询关卡失败: {e}")
            return None
    
    def update_level(self, level_id: str, **kwargs) -> bool:
        """更新关卡信息 (支持坐标更新)"""
        if not kwargs:
            return False
        
        try:
            set_clauses = []
            values = []
            
            for field, value in kwargs.items():
                if field == 'name':
                    set_clauses.append("name = ?")
                    values.append(value)
                elif field == 'pass_condition':
                    set_clauses.append("pass_condition = ?")
                    values.append(json.dumps(value, ensure_ascii=False))
                elif field == 'meta_json':
                    set_clauses.append("meta_json = ?")
                    values.append(json.dumps(value, ensure_ascii=False) if value else None)
                elif field in ['x', 'y']:
                    set_clauses.append(f"{field} = ?")
                    values.append(value)
            
            if not set_clauses:
                return False
            
            # 添加更新时间
            set_clauses.append("updated_at = ?")
            values.append(self._get_current_timestamp())
            values.append(level_id)
            
            sql = f"UPDATE levels SET {', '.join(set_clauses)} WHERE id = ?"
            
            with self.get_connection() as conn:
                cursor = conn.execute(sql, values)
                conn.commit()
                
                if cursor.rowcount > 0:
                    self.logger.info(f"更新关卡成功: {level_id}")
                    return True
                else:
                    self.logger.warning(f"关卡不存在: {level_id}")
                    return False
        except Exception as e:
            self.logger.error(f"更新关卡失败: {e}")
            return False
    
    def delete_level(self, level_id: str) -> bool:
        """删除关卡 (级联删除题目)"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("DELETE FROM levels WHERE id = ?", (level_id,))
                conn.commit()
                
                if cursor.rowcount > 0:
                    self.logger.info(f"删除关卡成功: {level_id}")
                    return True
                else:
                    self.logger.warning(f"关卡不存在: {level_id}")
                    return False
        except Exception as e:
            self.logger.error(f"删除关卡失败: {e}")
            return False
    
    # ========== 题目相关操作 ==========
    
    def insert_question(self, level_id: str, stem: str, content_json: Dict, 
                       answer_json: Dict, score: int, created_by: str = None) -> str:
        """插入题目记录"""
        try:
            question_id = self._generate_id()
            current_time = self._get_current_timestamp()
            
            with self.get_connection() as conn:
                conn.execute("""
                    INSERT INTO questions (id, level_id, stem, content_json, answer_json, score, created_by, created_at)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """, (question_id, level_id, stem, 
                     json.dumps(content_json, ensure_ascii=False),
                     json.dumps(answer_json, ensure_ascii=False),
                     score, created_by, current_time))
                
                conn.commit()
                self.logger.info(f"成功插入题目: {stem[:20]}..., ID: {question_id}")
                return question_id
        except Exception as e:
            self.logger.error(f"插入题目失败: {e}")
            raise
    
    def get_question_by_id(self, question_id: str) -> Optional[Dict]:
        """根据ID查询题目"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM questions WHERE id = ?", (question_id,))
                row = cursor.fetchone()
                if row:
                    result = dict(row)
                    # 解析JSON字段
                    result['content_json'] = json.loads(result['content_json'])
                    result['answer_json'] = json.loads(result['answer_json'])
                    return result
                return None
        except Exception as e:
            self.logger.error(f"查询题目失败: {e}")
            return None
    
    def get_questions_by_level_id(self, level_id: str) -> List[Dict]:
        """查询关卡下的所有题目"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("SELECT * FROM questions WHERE level_id = ? ORDER BY created_at", (level_id,))
                results = []
                for row in cursor.fetchall():
                    result = dict(row)
                    # 解析JSON字段
                    result['content_json'] = json.loads(result['content_json'])
                    result['answer_json'] = json.loads(result['answer_json'])
                    results.append(result)
                return results
        except Exception as e:
            self.logger.error(f"查询题目列表失败: {e}")
            return []
    
    def update_question(self, question_id: str, **kwargs) -> bool:
        """更新题目信息"""
        if not kwargs:
            return False
        
        try:
            set_clauses = []
            values = []
            
            for field, value in kwargs.items():
                if field == 'stem':
                    set_clauses.append("stem = ?")
                    values.append(value)
                elif field == 'content_json':
                    set_clauses.append("content_json = ?")
                    values.append(json.dumps(value, ensure_ascii=False))
                elif field == 'answer_json':
                    set_clauses.append("answer_json = ?")
                    values.append(json.dumps(value, ensure_ascii=False))
                elif field == 'score':
                    set_clauses.append("score = ?")
                    values.append(value)
                elif field == 'created_by':
                    set_clauses.append("created_by = ?")
                    values.append(value)
            
            if not set_clauses:
                return False
            
            values.append(question_id)
            sql = f"UPDATE questions SET {', '.join(set_clauses)} WHERE id = ?"
            
            with self.get_connection() as conn:
                cursor = conn.execute(sql, values)
                conn.commit()
                
                if cursor.rowcount > 0:
                    self.logger.info(f"更新题目成功: {question_id}")
                    return True
                else:
                    self.logger.warning(f"题目不存在: {question_id}")
                    return False
        except Exception as e:
            self.logger.error(f"更新题目失败: {e}")
            return False
    
    def delete_question(self, question_id: str) -> bool:
        """删除题目"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("DELETE FROM questions WHERE id = ?", (question_id,))
                conn.commit()
                
                if cursor.rowcount > 0:
                    self.logger.info(f"删除题目成功: {question_id}")
                    return True
                else:
                    self.logger.warning(f"题目不存在: {question_id}")
                    return False
        except Exception as e:
            self.logger.error(f"删除题目失败: {e}")
            return False
    
    # ========== 路线图相关操作 ==========
    
    def insert_roadmap_node(self, subject_id: str, level_id: str, parent_id: str = None,
                           sort_order: int = 1, path: str = "", depth: int = 0) -> str:
        """插入路线图节点"""
        try:
            node_id = self._generate_id()
            
            with self.get_connection() as conn:
                conn.execute("""
                    INSERT INTO roadmap_nodes (id, subject_id, level_id, parent_id, sort_order, path, depth)
                    VALUES (?, ?, ?, ?, ?, ?, ?)
                """, (node_id, subject_id, level_id, parent_id, sort_order, path, depth))
                
                conn.commit()
                self.logger.info(f"成功插入路线图节点: {node_id}")
                return node_id
        except Exception as e:
            self.logger.error(f"插入路线图节点失败: {e}")
            raise
    
    def get_roadmap_nodes_by_subject(self, subject_id: str) -> List[Dict]:
        """查询学科的路线图节点"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute("""
                    SELECT rn.*, l.name as level_name, l.x, l.y
                    FROM roadmap_nodes rn
                    JOIN levels l ON rn.level_id = l.id
                    WHERE rn.subject_id = ?
                    ORDER BY rn.sort_order
                """, (subject_id,))
                return [dict(row) for row in cursor.fetchall()]
        except Exception as e:
            self.logger.error(f"查询路线图节点失败: {e}")
            return []
    
    # ========== 统计相关操作 ==========
    
    def get_system_stats(self) -> Dict[str, Any]:
        """获取系统统计信息"""
        try:
            with self.get_connection() as conn:
                stats = {}
                
                # 学科统计
                cursor = conn.execute("SELECT COUNT(*) as count FROM subjects")
                stats['total_subjects'] = cursor.fetchone()['count']
                
                # 论文统计
                cursor = conn.execute("SELECT COUNT(*) as count FROM papers")
                stats['total_papers'] = cursor.fetchone()['count']
                
                # 关卡统计
                cursor = conn.execute("SELECT COUNT(*) as count FROM levels")
                stats['total_levels'] = cursor.fetchone()['count']
                
                # 题目统计
                cursor = conn.execute("SELECT COUNT(*) as count FROM questions")
                stats['total_questions'] = cursor.fetchone()['count']
                
                # 用户统计
                cursor = conn.execute("SELECT COUNT(*) as count FROM users")
                stats['total_users'] = cursor.fetchone()['count']
                
                # 路线图节点统计
                cursor = conn.execute("SELECT COUNT(*) as count FROM roadmap_nodes")
                stats['total_roadmap_nodes'] = cursor.fetchone()['count']
                
                # 平均每关卡题目数
                cursor = conn.execute("""
                    SELECT AVG(question_count) as avg_questions
                    FROM (
                        SELECT COUNT(*) as question_count 
                        FROM questions 
                        GROUP BY level_id
                    )
                """)
                avg_result = cursor.fetchone()
                stats['avg_questions_per_level'] = round(avg_result['avg_questions'] or 0, 2)
                
                # 题目分值统计
                cursor = conn.execute("SELECT AVG(score) as avg_score, MAX(score) as max_score, MIN(score) as min_score FROM questions")
                score_stats = cursor.fetchone()
                stats['avg_question_score'] = round(score_stats['avg_score'] or 0, 2)
                stats['max_question_score'] = score_stats['max_score'] or 0
                stats['min_question_score'] = score_stats['min_score'] or 0
                
                return stats
        except Exception as e:
            self.logger.error(f"获取统计信息失败: {e}")
            return {}
    
    def get_paper_stats(self) -> Dict[str, Any]:
        """获取论文统计信息 (为了兼容性保留)"""
        stats = self.get_system_stats()
        return {
            'total_papers': stats.get('total_papers', 0),
            'total_levels': stats.get('total_levels', 0),
            'total_questions': stats.get('total_questions', 0)
        }
    
    # ========== 辅助方法 ==========
    
    def get_paper_with_level_and_questions(self, paper_id: str) -> Optional[Dict]:
        """获取论文的完整信息（包含关卡和题目）"""
        paper = self.get_paper_by_id(paper_id)
        if not paper:
            return None
        
        level = self.get_level_by_paper_id(paper_id)
        if level:
            questions = self.get_questions_by_level_id(level['id'])
            level['questions'] = questions
            paper['level'] = level
        
        return paper
    
    def search_papers_by_title(self, keyword: str) -> List[Dict]:
        """根据标题关键词搜索论文"""
        try:
            with self.get_connection() as conn:
                cursor = conn.execute(
                    "SELECT * FROM papers WHERE title LIKE ? ORDER BY created_at DESC", 
                    (f"%{keyword}%",)
                )
                return [dict(row) for row in cursor.fetchall()]
        except Exception as e:
            self.logger.error(f"搜索论文失败: {e}")
            return []
    
    def close(self):
        """清理资源"""
        pass 