# Advanced Guides

Thư mục này chứa các hướng dẫn nâng cao cho hệ thống ABAC.

## 📚 Tài Liệu Có Sẵn

Hiện tại thư mục này đang được tái cấu trúc để loại bỏ các tài liệu version cũ và tập trung vào documentation hiện tại.

## 🎯 Mục Đích

Các tài liệu trong folder này được thiết kế cho:

### 👨‍💼 **Enterprise Architects**
- Hiểu các use cases phức tạp
- Đánh giá khả năng áp dụng cho tổ chức
- Planning cho multi-tenant architectures

### 👨‍💻 **Senior Developers**
- Deep dive vào technical implementation
- Performance optimization strategies
- Advanced pattern matching techniques

### 🔒 **Security Engineers**
- Compliance requirements (HIPAA, SOX, PCI-DSS)
- Security boundary analysis
- Audit trail implementation

### 🏗️ **System Designers**
- Hierarchical resource modeling
- Scalability considerations
- Integration patterns

## 📋 Prerequisites

Trước khi đọc các tài liệu nâng cao này, bạn nên đã nắm vững:

1. **Basic ABAC Concepts** - Đọc README.md chính
2. **Resource Field Documentation** - Đọc `../RESOURCE_FIELD_DOCUMENTATION.md`
3. **Action Field Documentation** - Đọc `../ACTION_FIELD_DOCUMENTATION.md`
4. **Condition Field Guide** - Đọc `../CONDITION_FIELD_GUIDE.md`
5. **Hierarchical Resource Guide** - Đọc `../HIERARCHICAL_RESOURCE_GUIDE.md`

## 🚀 Demo Code

Demo code cho hierarchical + extended format đã được chuyển sang thư mục `examples/`:

```bash
# Chạy demo code
cd ../../examples/
go run hierarchical_extended_demo.go
```

## 🤝 Contributing

Khi thêm tài liệu nâng cao mới:

1. **Follow naming convention**: `TOPIC_GUIDE.md`
2. **Include practical examples** với real-world scenarios
3. **Add performance considerations** nếu applicable
4. **Update this README** với link và description
5. **Cross-reference** với basic documentation

---

*Thư mục này đang được tái cấu trúc để cung cấp documentation ngắn gọn, đầy đủ và dễ hiểu cho người mới.*