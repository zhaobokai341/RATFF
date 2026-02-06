import json
import os

class LanguagePack:
    def __init__(self, file: str, language: str):
        self.file = file
        self.language = language
        self.file = os.path.join("../lang_pack/", self.file)
    
    def load(self):
        with open(self.file, 'r', encoding="utf-8") as f:
            self.data = json.load(f)
    
    def g(self, key: str) -> str:
        return self.data[key][self.language]
