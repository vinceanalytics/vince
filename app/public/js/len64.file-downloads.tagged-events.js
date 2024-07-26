!function(){"use strict";var r=window.location,i=window.document,o=i.currentScript,p=o.getAttribute("data-api")||new URL(o.src).origin+"/api/event";function l(t,e){t&&console.warn("Ignoring Event: "+t),e&&e.callback&&e.callback()}function t(t,e){if(/^localhost$|^127(\.[0-9]+){0,2}\.[0-9]+$|^\[::1?\]$/.test(r.hostname)||"file:"===r.protocol)return l("localhost",e);if((window._phantom||window.__nightmare||window.navigator.webdriver||window.Cypress)&&!window.__plausible)return l(null,e);try{if("true"===window.localStorage.plausible_ignore)return l("localStorage flag",e)}catch(t){}var a={},n=(a.n=t,a.u=r.href,a.d=o.getAttribute("data-domain"),a.r=i.referrer||null,e&&e.meta&&(a.m=JSON.stringify(e.meta)),e&&e.props&&(a.p=e.props),new XMLHttpRequest);n.open("POST",p,!0),n.setRequestHeader("Content-Type","text/plain"),n.send(JSON.stringify(a)),n.onreadystatechange=function(){4===n.readyState&&e&&e.callback&&e.callback({status:n.status})}}var e=window.plausible&&window.plausible.q||[];window.plausible=t;for(var a,n=0;n<e.length;n++)t.apply(this,e[n]);function s(){a!==r.pathname&&(a=r.pathname,t("pageview"))}var u,c=window.history;function f(t){return t&&t.tagName&&"a"===t.tagName.toLowerCase()}c.pushState&&(u=c.pushState,c.pushState=function(){u.apply(this,arguments),s()},window.addEventListener("popstate",s)),"prerender"===i.visibilityState?i.addEventListener("visibilitychange",function(){a||"visible"!==i.visibilityState||s()}):s();var d=1;function m(t){var e,a,n,r;if("auxclick"!==t.type||t.button===d)return e=function(t){for(;t&&(void 0===t.tagName||!f(t)||!t.href);)t=t.parentNode;return t}(t.target),a=e&&e.href&&e.href.split("?")[0],!function t(e,a){if(!e||y<a)return!1;if(L(e))return!0;return t(e.parentNode,a+1)}(e,0)&&(n=a)&&(r=n.split(".").pop(),b.some(function(t){return t===r}))?v(t,e,{name:"File Download",props:{url:a}}):void 0}function v(t,e,a){var n,r=!1;function i(){r||(r=!0,window.location=e.href)}!function(t,e){if(!t.defaultPrevented)return e=!e.target||e.target.match(/^_(self|parent|top)$/i),t=!(t.ctrlKey||t.metaKey||t.shiftKey)&&"click"===t.type,e&&t}(t,e)?(n={props:a.props},plausible(a.name,n)):(n={props:a.props,callback:i},plausible(a.name,n),setTimeout(i,5e3),t.preventDefault())}i.addEventListener("click",m),i.addEventListener("auxclick",m);var c=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],w=o.getAttribute("file-types"),g=o.getAttribute("add-file-types"),b=w&&w.split(",")||g&&g.split(",").concat(c)||c;function h(t){var t=L(t)?t:t&&t.parentNode,e={name:null,props:{}},a=t&&t.classList;if(a)for(var n=0;n<a.length;n++){var r,i=a.item(n).match(/plausible-event-(.+)(=|--)(.+)/);i&&(r=i[1],i=i[3].replace(/\+/g," "),"name"==r.toLowerCase()?e.name=i:e.props[r]=i)}return e}var y=3;function k(t){if("auxclick"!==t.type||t.button===d){for(var e,a,n,r,i=t.target,o=0;o<=y&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;f(i)&&(e=i),L(i)&&(a=i),i=i.parentNode}a&&(r=h(a),e?(r.props.url=e.href,v(t,e,r)):((t={}).props=r.props,plausible(r.name,t)))}}function L(t){var e=t&&t.classList;if(e)for(var a=0;a<e.length;a++)if(e.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}i.addEventListener("submit",function(t){var e,a=t.target,n=h(a);function r(){e||(e=!0,a.submit())}n.name&&(t.preventDefault(),e=!1,setTimeout(r,5e3),t={props:n.props,callback:r},plausible(n.name,t))}),i.addEventListener("click",k),i.addEventListener("auxclick",k)}();