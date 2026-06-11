from setuptools import setup, find_packages

setup(
    name="lychee-python",
    version="0.1.0",
    description="Official Python SDK for the Lychee local LLM runtime",
    long_description=open("README.md").read(),
    long_description_content_type="text/markdown",
    author="MD-Mushfiqur123",
    url="https://github.com/MD-Mushfiqur123/lychee",
    packages=find_packages(),
    python_requires=">=3.9",
    install_requires=[],  # Zero dependencies — uses stdlib only
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Scientific/Engineering :: Artificial Intelligence",
    ],
    keywords=["llm", "ai", "local-ai", "ollama", "anthropic", "openai"],
)
