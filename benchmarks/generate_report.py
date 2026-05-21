#!/usr/bin/env python3
"""
Benchmark Report Generator
Reads k6 JSON output and generates a markdown report
"""

import json
import os
import sys
from datetime import datetime
from pathlib import Path


def load_k6_summary(filepath):
    """Load k6 summary JSON file"""
    try:
        with open(filepath) as f:
            return json.load(f)
    except Exception as e:
        print(f"Error loading {filepath}: {e}")
        return None


def format_duration(ms):
    """Format duration in human readable format"""
    if ms < 1000:
        return f"{ms:.1f}ms"
    elif ms < 60000:
        return f"{ms/1000:.2f}s"
    else:
        return f"{ms/60000:.2f}m"


def generate_report(results_dir):
    """Generate benchmark report from k6 results"""
    results_path = Path(results_dir)
    
    report = []
    report.append("# Industrial AI Platform - Performance Benchmark Report")
    report.append("")
    report.append(f"**Generated**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    report.append(f"**Test Directory**: `{results_dir}`")
    report.append("")
    
    # Find all summary files
    summary_files = list(results_path.glob("*_summary.json"))
    
    if not summary_files:
        report.append("⚠️ No benchmark results found")
        return "\n".join(report)
    
    # Overview table
    report.append("## 📊 Benchmark Summary")
    report.append("")
    report.append("| Test | Duration | Requests | P95 Latency | Error Rate | Pass/Fail |")
    report.append("|------|----------|----------|-------------|------------|-----------|")
    
    overall_pass = True
    
    for summary_file in summary_files:
        summary = load_k6_summary(summary_file)
        if not summary:
            continue
        
        test_name = summary_file.stem.replace("_summary", "")
        
        # Extract metrics
        root_group = summary.get("root_group", {})
        metrics = summary.get("metrics", {})
        
        # HTTP request metrics
        http_reqs = metrics.get("http_reqs", {})
        total_requests = http_reqs.get("value", 0)
        
        http_req_duration = metrics.get("http_req_duration", {})
        p95 = http_req_duration.get("values", {}).get("p(95)", 0)
        
        http_req_failed = metrics.get("http_req_failed", {})
        fail_rate = http_req_failed.get("values", {}).get("rate", 0)
        
        # Test state
        test_state = summary.get("state", {})
        test_duration_ms = test_state.get("testRunDurationMs", 0)
        
        # Check thresholds
        thresholds_pass = summary.get("thresholds", {})
        passed = thresholds_pass.get("http_req_duration", {}).get("ok", True) and \
                 thresholds_pass.get("http_req_failed", {}).get("ok", True)
        
        status = "✅ Pass" if passed else "❌ Fail"
        if not passed:
            overall_pass = False
        
        report.append(f"| {test_name} | {format_duration(test_duration_ms)} | {total_requests} | {format_duration(p95)} | {fail_rate:.2%} | {status} |")
    
    report.append("")
    
    # Overall result
    report.append(f"**Overall Result**: {'✅ All tests passed' if overall_pass else '❌ Some tests failed'}")
    report.append("")
    
    # Detailed metrics for each test
    report.append("## 📈 Detailed Metrics")
    report.append("")
    
    for summary_file in summary_files:
        summary = load_k6_summary(summary_file)
        if not summary:
            continue
        
        test_name = summary_file.stem.replace("_summary", "")
        report.append(f"### {test_name}")
        report.append("")
        
        metrics = summary.get("metrics", {})
        
        # HTTP metrics
        report.append("**HTTP Metrics**:")
        report.append("")
        report.append("| Metric | Min | Avg | Med | P90 | P95 | Max |")
        report.append("|--------|-----|-----|-----|-----|-----|-----|")
        
        for metric_name in ["http_req_duration", "http_req_connecting", "http_req_receiving"]:
            metric = metrics.get(metric_name, {})
            values = metric.get("values", {})
            if values:
                report.append(f"| {metric_name} | {format_duration(values.get('min', 0))} | {format_duration(values.get('avg', 0))} | {format_duration(values.get('med', 0))} | {format_duration(values.get('p(90)', 0))} | {format_duration(values.get('p(95)', 0))} | {format_duration(values.get('max', 0))} |")
        
        report.append("")
        
        # Request counts
        http_reqs = metrics.get("http_reqs", {})
        if http_reqs:
            report.append(f"- **Total Requests**: {http_reqs.get('value', 0)}")
            report.append(f"- **Requests/sec**: {http_reqs.get('rate', 0):.2f}")
        
        # Error rate
        http_req_failed = metrics.get("http_req_failed", {})
        if http_req_failed:
            report.append(f"- **Error Rate**: {http_req_failed.get('values', {}).get('rate', 0):.2%}")
        
        report.append("")
    
    # Recommendations
    report.append("## 🔧 Recommendations")
    report.append("")
    
    # Analyze results and provide recommendations
    recommendations = []
    
    for summary_file in summary_files:
        summary = load_k6_summary(summary_file)
        if not summary:
            continue
        
        metrics = summary.get("metrics", {})
        p95 = metrics.get("http_req_duration", {}).get("values", {}).get("p(95)", 0)
        
        if p95 > 1000:
            recommendations.append(f"- High latency detected (>1s). Consider adding caching or optimizing queries.")
        
        fail_rate = metrics.get("http_req_failed", {}).get("values", {}).get("rate", 0)
        if fail_rate > 0.01:
            recommendations.append(f"- Error rate above 1%. Check for rate limiting or connection issues.")
    
    if recommendations:
        report.extend(recommendations)
    else:
        report.append("- Performance within acceptable thresholds. No immediate optimizations required.")
    
    report.append("")
    
    # Footer
    report.append("---")
    report.append("")
    report.append("*Report generated by Industrial AI Platform Benchmark Suite*")
    
    return "\n".join(report)


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 generate_report.py <results_dir>")
        sys.exit(1)
    
    results_dir = sys.argv[1]
    report = generate_report(results_dir)
    print(report)