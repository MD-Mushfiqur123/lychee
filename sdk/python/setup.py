from setuptools import setup

setup(
    name="lychee",
    version="0.2.0",
    description="Python client for Lychee — universal LLM runtime",
    long_description=open("README.md", encoding="utf-8").read(),
    long_description_content_type="text/markdown",
    author="Lychee Tech",
    url="https://github.com/MD-Mushfiqur123/lychee",
    py_modules=["lychee"],
    install_requires=["requests"],
    python_requires=">=3.8",
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
    ],
    keywords=["lychee", "llm", "ai", "local-ai"],
)
