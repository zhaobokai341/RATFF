import json
import os

class LanguagePack:
    def __init__(self, file, language):
        self.file = file
        self.language = language
        self.file = os.path.join("../lang_pack", self.file)
    
    def load(self):
        with open(self.file, 'r') as f:
            self.data = json.load(f)
    
    def g(self, key):
        return self.data[key][self.language]
