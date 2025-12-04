;(function (global, factory) {
   typeof exports === 'object' && typeof module !== 'undefined'
       && typeof require === 'function' ? factory(require('./.')) :
   typeof define === 'function' && define.amd ? define(['./.'], factory) :
   factory(global.moment)
}(this, function (moment) { 'use strict';
