!function(){"use strict";var o=window.location,p=window.document,l=p.getElementById("plausible"),s=l.getAttribute("data-api")||(g=(g=l).src.split("/"),c=g[0],g=g[2],c+"//"+g+"/api/event");function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(o.hostname)||"file:"===o.protocol)return u("localhost",t);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return u(null,t);try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var a=l&&l.getAttribute("data-include"),n=l&&l.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(r),n=n&&n.split(",").some(r);if(!a||n)return u("exclusion rule",t)}function r(e){return o.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},i=(a.n=e,a.u=o.href,a.d=l.getAttribute("data-domain"),a.r=p.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,n=0;n<t.length;n++)e.apply(this,t[n]);function r(){a!==o.pathname&&(a=o.pathname,e("pageview"))}var i,c=window.history;function f(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}c.pushState&&(i=c.pushState,c.pushState=function(){i.apply(this,arguments),r()},window.addEventListener("popstate",r)),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){a||"visible"!==p.visibilityState||r()}):r();var d=1;function m(e){var t,a,n,r;if("auxclick"!==e.type||e.button===d)return t=function(e){for(;e&&(void 0===e.tagName||!f(e)||!e.href);)e=e.parentNode;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],!function e(t,a){if(!t||k<a)return!1;if(L(t))return!0;return e(t.parentNode,a+1)}(t,0)&&(n=a)&&(r=n.split(".").pop(),h.some(function(e){return e===r}))?v(e,t,{name:"File Download",props:{url:a}}):void 0}function v(e,t,a){var n,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(n={props:a.props},plausible(a.name,n)):(n={props:a.props,callback:i},plausible(a.name,n),setTimeout(i,5e3),e.preventDefault())}p.addEventListener("click",m),p.addEventListener("auxclick",m);var g=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],w=l.getAttribute("file-types"),b=l.getAttribute("add-file-types"),h=w&&w.split(",")||b&&b.split(",").concat(g)||g;function y(e){var e=L(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var n=0;n<a.length;n++){var r,i=a.item(n).match(/plausible-event-(.+)(=|--)(.+)/);i&&(r=i[1],i=i[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i)}return t}var k=3;function x(e){if("auxclick"!==e.type||e.button===d){for(var t,a,n,r,i=e.target,o=0;o<=k&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;f(i)&&(t=i),L(i)&&(a=i),i=i.parentNode}a&&(r=y(a),t?(r.props.url=t.href,v(e,t,r)):((e={}).props=r.props,plausible(r.name,e)))}}function L(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}p.addEventListener("submit",function(e){var t,a=e.target,n=y(a);function r(){t||(t=!0,a.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),e={props:n.props,callback:r},plausible(n.name,e))}),p.addEventListener("click",x),p.addEventListener("auxclick",x)}();